package recognizer

import (
	"fmt"

	"mini-tmk-agent/internal/config"
)

// DefaultLanguage 默认识别语言
const DefaultLanguage = "zh"

// New 创建默认的 ASR 识别器（使用百炼平台）
func New(language string) (Recognizer, error) {
	cfg := config.GetInstance()

	if language == "" {
		language = DefaultLanguage
	}

	// 使用百炼平台 ASR（统一使用百炼 API Key）
	apiKey, ok := cfg.GetBaiwanAPIKey()
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("百炼平台 API Key 未配置，请先运行: mini-tmk-agent config set-baiwan-key <api-key>")
	}

	return newBaiwanASR(apiKey, language)
}

// NewWithAPIKey 使用指定的百炼 API Key 创建识别器
func NewWithAPIKey(apiKey, language string) (Recognizer, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("百炼平台 API Key 不能为空")
	}

	if language == "" {
		language = DefaultLanguage
	}

	return newBaiwanASR(apiKey, language)
}

// MustNew 创建识别器，失败时 panic
func MustNew(language string) Recognizer {
	r, err := New(language)
	if err != nil {
		panic(err)
	}
	return r
}

// IsAPIKeyConfigured 检查百炼平台 API Key 是否已配置
func IsAPIKeyConfigured() bool {
	cfg := config.GetInstance()
	apiKey, ok := cfg.GetBaiwanAPIKey()
	return ok && apiKey != ""
}

// GetCurrentAPIKeySource 返回当前百炼 API Key 来源
func GetCurrentAPIKeySource() (source string, maskedKey string) {
	cfg := config.GetInstance()
	apiKey, ok := cfg.GetBaiwanAPIKey()
	if !ok || apiKey == "" {
		return "未配置", ""
	}
	return "用户配置", MaskKey(apiKey)
}

// MaskKey 隐藏密钥中间部分
func MaskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}
