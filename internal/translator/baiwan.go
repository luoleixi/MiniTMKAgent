package translator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"mini-tmk-agent/internal/config"
	"mini-tmk-agent/internal/utils"
)

// BaiwanTranslator 百炼平台翻译实现
type BaiwanTranslator struct {
	apiKey string
	client *http.Client
}

// NewBaiwanTranslator 创建百炼平台翻译器
func NewBaiwanTranslator() (*BaiwanTranslator, error) {
	cfg := config.GetInstance()
	apiKey, ok := cfg.GetBaiwanAPIKey()
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("百炼平台 API Key 未配置")
	}

	return &BaiwanTranslator{
		apiKey: apiKey,
		client: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Translate 执行翻译
func (t *BaiwanTranslator) Translate(sourceLang, targetLang, text string) (string, error) {
	if text == "" {
		return "", nil
	}

	// 调用百炼平台翻译
	return t.callBaiwanTranslate(sourceLang, targetLang, text)
}

// callBaiwanTranslate 调用百炼平台翻译 API
func (t *BaiwanTranslator) callBaiwanTranslate(sourceLang, targetLang, text string) (string, error) {
	// 使用百炼平台统一接口
	url := "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"

	// 构建翻译提示
	prompt := fmt.Sprintf("请将以下%s文本翻译成%s：%s", getLangName(sourceLang), getLangName(targetLang), text)

	reqBody := map[string]interface{}{
		"model": "qwen-turbo",
		"input": map[string]interface{}{
			"messages": []map[string]string{
				{
					"role":    "system",
					"content": "你是一个翻译助手，只返回翻译结果，不要添加任何解释。",
				},
				{
					"role":    "user",
					"content": prompt,
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return text, nil // 失败返回原文
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return text, nil
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.apiKey)

	resp, err := t.client.Do(req)
	if err != nil {
		return text, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return text, nil
	}

	// 记录调试信息
	utils.Debugf("翻译API响应: %s", string(body))

	// 解析响应 (百炼平台格式)
	var result struct {
		Output struct {
			Text         string `json:"text"`
			FinishReason string `json:"finish_reason"`
		} `json:"output"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		// 解析失败返回原文
		utils.Debugf("翻译解析失败: %v", err)
		return text, nil
	}

	if result.Output.Text != "" {
		return result.Output.Text, nil
	}

	// 失败返回原文
	return text, nil
}

// getLangName 获取语言名称
func getLangName(code string) string {
	names := map[string]string{
		"zh": "中文",
		"en": "英文",
		"ja": "日文",
		"ko": "韩文",
		"fr": "法文",
		"de": "德文",
		"es": "西班牙文",
		"ru": "俄文",
	}
	if name, ok := names[code]; ok {
		return name
	}
	return code
}

// IsConfigured 检查是否已配置
func (t *BaiwanTranslator) IsConfigured() bool {
	return t.apiKey != ""
}
