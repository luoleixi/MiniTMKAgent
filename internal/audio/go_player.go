package audio

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
)

var (
	// 全局单例
	globalPlayer     *GoPlayer
	globalPlayerOnce sync.Once
	globalPlayerErr  error
)

// GoPlayer 纯 Go 实现的 MP3 播放器（无需外部依赖）
type GoPlayer struct {
	context *oto.Context
	ready   chan struct{}
}

// GetGoPlayer 获取全局 GoPlayer 单例（oto.Context 只能创建一次）
func GetGoPlayer() (*GoPlayer, error) {
	globalPlayerOnce.Do(func() {
		globalPlayer, globalPlayerErr = newGoPlayer()
	})
	return globalPlayer, globalPlayerErr
}

// newGoPlayer 创建纯 Go MP3 播放器
// 使用 44100Hz 作为目标采样率（CD 音质，最通用）
func newGoPlayer() (*GoPlayer, error) {
	op := &oto.NewContextOptions{
		SampleRate:   44100, // CD 音质，最通用的采样率
		ChannelCount: 2,     // 立体声
		Format:       oto.FormatSignedInt16LE,
	}

	context, ready, err := oto.NewContext(op)
	if err != nil {
		return nil, fmt.Errorf("初始化音频上下文失败: %w", err)
	}

	return &GoPlayer{
		context: context,
		ready:   ready,
	}, nil
}

// PlayMP3Data 播放 MP3 数据
func (p *GoPlayer) PlayMP3Data(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("音频数据为空")
	}

	// 等待音频上下文就绪
	<-p.ready

	// 解码 MP3 为 PCM
	pcmData, sampleRate, channels, err := decodeMP3(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("MP3 解码失败: %w", err)
	}

	// 重采样到 44100Hz 立体声（如果必要）
	if sampleRate != 44100 || channels != 2 {
		pcmData = resample(pcmData, sampleRate, 44100, channels, 2)
	}

	// 创建播放器并播放
	player := p.context.NewPlayer(bytes.NewReader(pcmData))
	player.Play()

	// 等待播放完成
	for player.IsPlaying() {
		time.Sleep(5 * time.Millisecond)
	}
	time.Sleep(50 * time.Millisecond) // 确保音频完全输出

	return nil
}

// decodeMP3 解码 MP3 为 PCM 数据
// 返回: PCM数据, 采样率, 声道数, 错误
func decodeMP3(r io.Reader) ([]byte, int, int, error) {
	decoder, err := mp3.NewDecoder(r)
	if err != nil {
		return nil, 0, 0, err
	}

	sampleRate := decoder.SampleRate()
	// go-mp3 默认返回立体声数据
	channels := 2

	// 读取所有 PCM 数据
	var pcmData []byte
	buf := make([]byte, 4096)
	for {
		n, err := decoder.Read(buf)
		if n > 0 {
			pcmData = append(pcmData, buf[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, 0, 0, err
		}
	}

	return pcmData, sampleRate, channels, nil
}

// resample 重采样 PCM 数据
// srcData: 源 PCM 数据 (16-bit 有符号, 小端序)
// srcRate: 源采样率
// dstRate: 目标采样率
// srcChannels: 源声道数
// dstChannels: 目标声道数
func resample(srcData []byte, srcRate, dstRate, srcChannels, dstChannels int) []byte {
	if srcRate == dstRate && srcChannels == dstChannels {
		return srcData
	}

	// 将字节转换为 int16 样本
	srcSamples := make([]int16, len(srcData)/2)
	for i := 0; i < len(srcSamples); i++ {
		srcSamples[i] = int16(srcData[i*2]) | (int16(srcData[i*2+1]) << 8)
	}

	// 计算目标样本数
	ratio := float64(dstRate) / float64(srcRate)
	dstSampleCount := int(float64(len(srcSamples)/srcChannels) * ratio)
	dstSamples := make([]int16, dstSampleCount*dstChannels)

	// 线性插值重采样
	for i := 0; i < dstSampleCount; i++ {
		srcPos := float64(i) / ratio
		srcIndex := int(srcPos)
		frac := srcPos - float64(srcIndex)

		for ch := 0; ch < dstChannels; ch++ {
			if ch < srcChannels {
				// 同声道插值
				idx1 := srcIndex*srcChannels + ch
				idx2 := (srcIndex + 1) * srcChannels + ch
				if idx2 >= len(srcSamples) {
					idx2 = idx1
				}

				s1 := float64(srcSamples[idx1])
				s2 := float64(srcSamples[idx2])
				value := s1 + (s2-s1)*frac
				dstSamples[i*dstChannels+ch] = int16(value)
			} else {
				// 单声道转立体声：复制左声道到右声道
				dstSamples[i*dstChannels+ch] = dstSamples[i*dstChannels]
			}
		}
	}

	// 转换回字节
	dstData := make([]byte, len(dstSamples)*2)
	for i, sample := range dstSamples {
		dstData[i*2] = byte(sample)
		dstData[i*2+1] = byte(sample >> 8)
	}

	return dstData
}

// PlayMP3Reader 从 io.Reader 播放 MP3
func (p *GoPlayer) PlayMP3Reader(r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("读取音频数据失败: %w", err)
	}
	return p.PlayMP3Data(data)
}

// Close 关闭播放器
func (p *GoPlayer) Close() error {
	return nil
}

// ResetGoPlayer 重置全局播放器
func ResetGoPlayer() {
	if globalPlayer != nil {
		globalPlayer.context.Suspend()
	}
	globalPlayer = nil
	globalPlayerOnce = sync.Once{}
}

// IsGoPlayerAvailable 检查纯 Go 播放器是否可用
func IsGoPlayerAvailable() bool {
	return true
}
