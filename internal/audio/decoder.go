package audio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/go-mp3"
)

// AudioDecoder 音频解码器接口
type AudioDecoder interface {
	Decode(reader io.Reader) ([]int16, error)
}

// DecoderFactory 解码器工厂
func NewDecoder(filename string) (AudioDecoder, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".pcm":
		return &PCMDecoder{}, nil
	case ".wav":
		return &WAVDecoder{}, nil
	case ".mp3":
		return &MP3Decoder{}, nil
	default:
		return nil, fmt.Errorf("不支持的音频格式: %s", ext)
	}
}

// PCMDecoder PCM裸数据解码器
type PCMDecoder struct{}

// Decode 解码PCM数据 (16bit, 单声道, 16kHz)
func (d *PCMDecoder) Decode(reader io.Reader) ([]int16, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// 确保数据长度为2的倍数 (16bit = 2 bytes)
	if len(data)%2 != 0 {
		data = data[:len(data)-1]
	}

	samples := make([]int16, len(data)/2)
	for i := 0; i < len(samples); i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(data[i*2 : (i*2)+2]))
	}

	return samples, nil
}

// WAVDecoder WAV文件解码器
type WAVDecoder struct{}

// WAVHeader WAV文件头
type WAVHeader struct {
	ChunkID       [4]byte
	ChunkSize     uint32
	Format        [4]byte
	Subchunk1ID   [4]byte
	Subchunk1Size uint32
	AudioFormat   uint16
	NumChannels   uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
}

// Decode 解码WAV文件
func (d *WAVDecoder) Decode(reader io.Reader) ([]int16, error) {
	// 读取WAV头部
	var header WAVHeader
	if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("读取WAV头部失败: %w", err)
	}

	// 验证WAV格式
	if string(header.ChunkID[:]) != "RIFF" || string(header.Format[:]) != "WAVE" {
		return nil, fmt.Errorf("无效的WAV文件")
	}

	// 查找data subchunk
	var subchunkID [4]byte
	var subchunkSize uint32

	for {
		if err := binary.Read(reader, binary.LittleEndian, &subchunkID); err != nil {
			return nil, fmt.Errorf("查找data chunk失败: %w", err)
		}
		if err := binary.Read(reader, binary.LittleEndian, &subchunkSize); err != nil {
			return nil, fmt.Errorf("读取chunk大小失败: %w", err)
		}

		if string(subchunkID[:]) == "data" {
			break
		}

		// 跳过当前chunk
		if _, err := io.CopyN(io.Discard, reader, int64(subchunkSize)); err != nil {
			return nil, fmt.Errorf("跳过chunk失败: %w", err)
		}
	}

	// 读取音频数据
	data := make([]byte, subchunkSize)
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, fmt.Errorf("读取音频数据失败: %w", err)
	}

	// 根据位深度转换样本
	var samples []int16
	bitsPerSample := int(header.BitsPerSample)
	if bitsPerSample == 0 {
		bitsPerSample = 16 // 默认值
	}

	switch bitsPerSample {
	case 8:
		// 8bit 无符号整数，需要转换为16bit有符号
		samples = make([]int16, len(data))
		for i := 0; i < len(data); i++ {
			// 8bit unsigned (0-255) -> 16bit signed (-32768 to 32767)
			samples[i] = int16(int(data[i])-128) * 256
		}
	case 16:
		// 16bit 有符号整数（小端序）
		samples = make([]int16, len(data)/2)
		for i := 0; i < len(samples); i++ {
			samples[i] = int16(binary.LittleEndian.Uint16(data[i*2 : (i*2)+2]))
		}
	case 24:
		// 24bit 有符号整数，需要转换为16bit
		samples = make([]int16, len(data)/3)
		for i := 0; i < len(samples); i++ {
			b0 := int32(data[i*3])
			b1 := int32(data[i*3+1])
			b2 := int32(data[i*3+2])
			// 24bit signed -> 16bit signed (取高16位)
			val := (b0 | (b1 << 8) | (b2 << 16))
			if val >= 8388608 { // 2^23
				val -= 16777216 // 2^24
			}
			samples[i] = int16(val >> 8)
		}
	case 32:
		// 32bit 浮点数或整数
		if header.AudioFormat == 3 { // IEEE float
			samples = make([]int16, len(data)/4)
			for i := 0; i < len(samples); i++ {
				bits := binary.LittleEndian.Uint32(data[i*4 : (i*4)+4])
				f := math.Float32frombits(bits)
				samples[i] = int16(f * 32767)
			}
		} else { // 32bit signed integer
			samples = make([]int16, len(data)/4)
			for i := 0; i < len(samples); i++ {
				val := int32(binary.LittleEndian.Uint32(data[i*4 : (i*4)+4]))
				samples[i] = int16(val >> 16)
			}
		}
	default:
		return nil, fmt.Errorf("不支持的位深度: %d bits", bitsPerSample)
	}

	// 如果需要，进行采样率转换和声道转换
	if header.SampleRate != 16000 || header.NumChannels != 1 {
		samples = convertAudio(samples, int(header.SampleRate), int(header.NumChannels), 16000, 1)
	}

	return samples, nil
}

// DecodeFile 解码音频文件
func DecodeFile(filename string) ([]int16, error) {
	decoder, err := NewDecoder(filename)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return decoder.Decode(file)
}

// MP3Decoder MP3解码器
type MP3Decoder struct{}

// Decode 解码MP3数据 (输出16kHz, 16bit, 单声道PCM)
func (d *MP3Decoder) Decode(reader io.Reader) ([]int16, error) {
	// 使用go-mp3解码
	decoder, err := mp3.NewDecoder(reader)
	if err != nil {
		return nil, fmt.Errorf("创建MP3解码器失败: %w", err)
	}

	// 获取MP3的采样率和声道数
	sampleRate := decoder.SampleRate()
	numChannels := 2 // MP3通常是立体声

	// 读取所有解码后的数据
	pcmData, err := io.ReadAll(decoder)
	if err != nil {
		return nil, fmt.Errorf("解码MP3失败: %w", err)
	}

	// MP3解码后通常是16bit PCM数据
	// 确保数据长度为2的倍数
	if len(pcmData)%2 != 0 {
		pcmData = pcmData[:len(pcmData)-1]
	}

	// 转换为int16样本
	samples := make([]int16, len(pcmData)/2)
	for i := 0; i < len(samples); i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(pcmData[i*2 : (i*2)+2]))
	}

	// 转换为16kHz单声道
	samples = convertAudio(samples, sampleRate, numChannels, 16000, 1)

	return samples, nil
}

// DecodeMP3Data 直接解码MP3字节数据
func DecodeMP3Data(data []byte) ([]int16, error) {
	return (&MP3Decoder{}).Decode(bytes.NewReader(data))
}

// convertAudio 转换音频采样率和声道数
func convertAudio(samples []int16, srcRate, srcChannels, dstRate, dstChannels int) []int16 {
	if srcRate == dstRate && srcChannels == dstChannels {
		return samples
	}

	// 先转换为单声道（如果需要）
	if srcChannels > 1 {
		monoSamples := make([]int16, len(samples)/srcChannels)
		for i := 0; i < len(monoSamples); i++ {
			sum := int32(0)
			for ch := 0; ch < srcChannels; ch++ {
				sum += int32(samples[i*srcChannels+ch])
			}
			monoSamples[i] = int16(sum / int32(srcChannels))
		}
		samples = monoSamples
	}

	// 重采样（如果需要）
	if srcRate != dstRate {
		// 简单的线性插值重采样
		ratio := float64(dstRate) / float64(srcRate)
		newLen := int(float64(len(samples)) * ratio)
		resampled := make([]int16, newLen)

		for i := 0; i < newLen; i++ {
			srcPos := float64(i) / ratio
			srcIdx := int(srcPos)
			frac := srcPos - float64(srcIdx)

			if srcIdx >= len(samples)-1 {
				resampled[i] = samples[len(samples)-1]
			} else {
				// 线性插值
				s1 := float64(samples[srcIdx])
				s2 := float64(samples[srcIdx+1])
				resampled[i] = int16(s1 + (s2-s1)*frac)
			}
		}
		samples = resampled
	}

	return samples
}
