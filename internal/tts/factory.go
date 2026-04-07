package tts

import (
	"fmt"
	"mini-tmk-agent/internal/config"
)

// New 创建默认的TTS客户端（使用百炼平台）
func New() (TTS, error) {
	cfg := config.GetInstance()
	apiKey, ok := cfg.GetBaiwanAPIKey()
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("百炼平台API Key未配置，请运行: mini-tmk-agent config set-baiwan-key <api-key>")
	}

	return NewBaiwanTTS(&Config{
		APIKey:     apiKey,
		SampleRate: DefaultSampleRate,
	})
}

// NewWithAPIKey 使用指定API Key创建TTS客户端
func NewWithAPIKey(apiKey string) (TTS, error) {
	return NewBaiwanTTS(&Config{
		APIKey:     apiKey,
		SampleRate: DefaultSampleRate,
	})
}

// IsConfigured 检查 TTS 配置是否已设置
func IsConfigured() bool {
	cfg := config.GetInstance()
	_, ok := cfg.GetBaiwanAPIKey()
	return ok
}
