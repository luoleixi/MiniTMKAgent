package translator

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"mini-tmk-agent/internal/config"
)

const (
	// DefaultAliyunTranslateEndpoint 阿里云翻译默认端点
	DefaultAliyunTranslateEndpoint = "https://mt.aliyuncs.com/"
)

// AliyunTranslateConfig 阿里云翻译配置
type AliyunTranslateConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	Endpoint        string
}

// AliyunTranslator 阿里云机器翻译实现
type AliyunTranslator struct {
	config *AliyunTranslateConfig
	client *http.Client
	mu     sync.RWMutex
}

// NewAliyunTranslator 创建阿里云翻译器
func NewAliyunTranslator(cfg *AliyunTranslateConfig) (*AliyunTranslator, error) {
	// 如果未提供配置，尝试从配置获取
	if cfg.AccessKeyID == "" {
		cfg.AccessKeyID = GetUserAliyunTranslateAccessKeyID()
	}
	if cfg.AccessKeySecret == "" {
		cfg.AccessKeySecret = GetUserAliyunTranslateAccessKeySecret()
	}

	// 如果仍然没有，尝试使用百炼 API Key（用于兼容）
	if cfg.AccessKeyID == "" {
		appCfg := config.GetInstance()
		if apiKey, ok := appCfg.GetBaiwanAPIKey(); ok && apiKey != "" {
			// 使用百炼 API Key 作为 AccessKey（简化处理）
			cfg.AccessKeyID = apiKey[:16]
			if len(apiKey) > 16 {
				cfg.AccessKeySecret = apiKey[16:]
			}
		}
	}

	if cfg.Endpoint == "" {
		cfg.Endpoint = DefaultAliyunTranslateEndpoint
	}

	return &AliyunTranslator{
		config: cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Translate 执行翻译
func (t *AliyunTranslator) Translate(sourceLang, targetLang, text string) (string, error) {
	if text == "" {
		return "", nil
	}

	// 如果未配置阿里云翻译，直接使用原文
	if t.config.AccessKeyID == "" || t.config.AccessKeySecret == "" {
		return text, nil
	}

	// 准备请求参数
	params := map[string]string{
		"Action":        "TranslateGeneral",
		"Format":        "JSON",
		"Version":       "2018-10-12",
		"AccessKeyId":   t.config.AccessKeyID,
		"SignatureMethod": "HMAC-SHA1",
		"Timestamp":     time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"SignatureVersion": "1.0",
		"SignatureNonce": fmt.Sprintf("%d", time.Now().UnixNano()),
		"SourceLanguage": sourceLang,
		"TargetLanguage": targetLang,
		"SourceText":    text,
	}

	// 计算签名
	signature := t.calculateSignature(params)
	params["Signature"] = signature

	// 构建请求 URL
	u, _ := url.Parse(t.config.Endpoint)
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	// 发送请求
	resp, err := t.client.Get(u.String())
	if err != nil {
		return "", fmt.Errorf("翻译请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	var result struct {
		Code      string `json:"Code"`
		Message   string `json:"Message"`
		Data struct {
			Translated string `json:"Translated"`
		} `json:"Data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != "200" {
		return "", fmt.Errorf("翻译失败: %s", result.Message)
	}

	return result.Data.Translated, nil
}

// calculateSignature 计算签名
func (t *AliyunTranslator) calculateSignature(params map[string]string) string {
	// 按参数名排序
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建规范化的查询字符串
	var canonicalQueryString []string
	for _, k := range keys {
		canonicalQueryString = append(canonicalQueryString,
			fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(params[k])))
	}

	// 构建签名字符串
	signString := "GET&%2F&" + url.QueryEscape(strings.Join(canonicalQueryString, "&"))

	// 计算签名
	signKey := t.config.AccessKeySecret + "&"
	h := hmac.New(sha1.New, []byte(signKey))
	h.Write([]byte(signString))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return signature
}

// IsConfigured 检查是否已配置
func (t *AliyunTranslator) IsConfigured() bool {
	return t.config.AccessKeyID != "" && t.config.AccessKeySecret != ""
}

// languageCodeMap 语言代码映射
var languageCodeMap = map[string]string{
	"zh": "zh",
	"en": "en",
	"ja": "ja",
	"ko": "ko",
	"fr": "fr",
	"de": "de",
	"es": "es",
	"ru": "ru",
	"ar": "ar",
	"pt": "pt",
	"it": "it",
	"nl": "nl",
	"pl": "pl",
	"tr": "tr",
	"th": "th",
	"vi": "vi",
	"id": "id",
	"ms": "ms",
}

// GetUserAliyunTranslateAccessKeyID 获取用户配置的 AccessKey ID
func GetUserAliyunTranslateAccessKeyID() string {
	// 简化：从配置获取百炼 API Key 作为兼容
	cfg := config.GetInstance()
	if apiKey, ok := cfg.GetBaiwanAPIKey(); ok && apiKey != "" {
		if len(apiKey) > 16 {
			return apiKey[:16]
		}
		return apiKey
	}
	return ""
}

// GetUserAliyunTranslateAccessKeySecret 获取用户配置的 AccessKey Secret
func GetUserAliyunTranslateAccessKeySecret() string {
	// 简化：从配置获取百炼 API Key 作为兼容
	cfg := config.GetInstance()
	if apiKey, ok := cfg.GetBaiwanAPIKey(); ok && apiKey != "" {
		if len(apiKey) > 16 {
			return apiKey[16:]
		}
	}
	return ""
}

// SaveAliyunTranslateConfig 保存阿里云翻译配置
func SaveAliyunTranslateConfig(accessKeyID, accessKeySecret string) error {
	// 不再单独保存阿里云翻译配置
	return nil
}

// IsAliyunTranslateConfigured 检查阿里云翻译配置是否已设置
func IsAliyunTranslateConfigured() bool {
	cfg := config.GetInstance()
	_, ok := cfg.GetBaiwanAPIKey()
	return ok
}

// GetCurrentAliyunTranslateConfigSource 返回当前配置来源
func GetCurrentAliyunTranslateConfigSource() (source, maskedID string) {
	cfg := config.GetInstance()
	apiKey, ok := cfg.GetBaiwanAPIKey()
	if !ok || apiKey == "" {
		return "未配置", ""
	}
	return "用户配置", maskKey(apiKey)
}

// maskKey 隐藏密钥中间部分
func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}
