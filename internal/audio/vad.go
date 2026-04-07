package audio

import (
	"sync"
	"time"
)

// VADConfig 语音活动检测配置
type VADConfig struct {
	SampleRate     int           // 采样率
	FrameSize      int           // 帧大小（样本数）
	EnergyThreshold float64      // 能量阈值
	ZCRThreshold   float64       // 过零率阈值
	MinSpeechDuration time.Duration // 最短语音时长
	MinSilenceDuration time.Duration // 最短静音时长
}

// DefaultVADConfig 默认VAD配置
func DefaultVADConfig() VADConfig {
	return VADConfig{
		SampleRate:         16000,
		FrameSize:          320,  // 20ms @ 16kHz
		EnergyThreshold:    0.01,
		ZCRThreshold:       0.1,
		MinSpeechDuration:  200 * time.Millisecond,
		MinSilenceDuration: 300 * time.Millisecond,
	}
}

// VAD 语音活动检测器
type VAD struct {
	config       VADConfig
	isSpeech     bool
	speechStart  time.Time
	silenceStart time.Time
	mu           sync.RWMutex
	onSpeechStart func()
	onSpeechEnd   func([]int16)
	buffer       *AudioBuffer
	speechBuffer []int16
}

// NewVAD 创建新的VAD检测器
func NewVAD(config VADConfig, onSpeechStart func(), onSpeechEnd func([]int16)) *VAD {
	return &VAD{
		config:        config,
		onSpeechStart: onSpeechStart,
		onSpeechEnd:   onSpeechEnd,
		buffer:        NewAudioBuffer(config.SampleRate * 10), // 10秒缓冲区
	}
}

// Process 处理音频帧
func (v *VAD) Process(frame []int16) {
	energy := v.calculateEnergy(frame)
	zcr := v.calculateZCR(frame)

	isSpeechFrame := energy > v.config.EnergyThreshold && zcr < v.config.ZCRThreshold

	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now()

	if isSpeechFrame {
		if !v.isSpeech {
			// 开始语音
			if v.silenceStart.IsZero() || now.Sub(v.silenceStart) < v.config.MinSilenceDuration {
				// 静音时间不够，继续等待
				v.buffer.Write(frame)
				return
			}
			v.isSpeech = true
			v.speechStart = now
			v.speechBuffer = v.buffer.ReadAll() // 获取预录的音频
			v.speechBuffer = append(v.speechBuffer, frame...)
			if v.onSpeechStart != nil {
				go v.onSpeechStart()
			}
		} else {
			// 继续语音
			v.speechBuffer = append(v.speechBuffer, frame...)
		}
		v.silenceStart = time.Time{}
	} else {
		if v.isSpeech {
			// 可能的语音结束
			if v.silenceStart.IsZero() {
				v.silenceStart = now
			}
			v.speechBuffer = append(v.speechBuffer, frame...)

			if now.Sub(v.silenceStart) >= v.config.MinSilenceDuration {
				// 确认语音结束
				v.isSpeech = false
				speechDuration := now.Sub(v.speechStart)
				if speechDuration >= v.config.MinSpeechDuration && v.onSpeechEnd != nil {
					go v.onSpeechEnd(v.speechBuffer)
				}
				v.speechBuffer = nil
			}
		} else {
			// 继续静音，保存到预录缓冲区
			v.buffer.Write(frame)
		}
	}
}

// IsSpeech 返回当前是否处于语音状态
func (v *VAD) IsSpeech() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.isSpeech
}

// calculateEnergy 计算帧能量
func (v *VAD) calculateEnergy(frame []int16) float64 {
	if len(frame) == 0 {
		return 0
	}

	var sum float64
	for _, sample := range frame {
		s := float64(sample) / 32768.0 // 归一化到[-1, 1]
		sum += s * s
	}
	return sum / float64(len(frame))
}

// calculateZCR 计算过零率
func (v *VAD) calculateZCR(frame []int16) float64 {
	if len(frame) < 2 {
		return 0
	}

	var zcr int
	for i := 1; i < len(frame); i++ {
		if (frame[i] >= 0 && frame[i-1] < 0) || (frame[i] < 0 && frame[i-1] >= 0) {
			zcr++
		}
	}

	return float64(zcr) / float64(len(frame)-1)
}

// Reset 重置VAD状态
func (v *VAD) Reset() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.isSpeech = false
	v.speechBuffer = nil
	v.buffer.Clear()
	v.silenceStart = time.Time{}
}
