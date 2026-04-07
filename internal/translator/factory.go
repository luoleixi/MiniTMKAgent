package translator
import "mini-tmk-agent/internal/config"

// DefaultLanguagePair 默认语言对
const (
	DefaultSourceLang = "zh"
	DefaultTargetLang = "en"
)

// New 创建默认的翻译器
// 优先使用百炼平台（统一API Key），其次使用阿里云翻译
func New() (Translator, error) {
	// 尝试使用百炼平台翻译（统一API Key）
	cfg := config.GetInstance()
	if apiKey, ok := cfg.GetBaiwanAPIKey(); ok && apiKey != "" {
		return NewBaiwanTranslator()
	}

	// 回退到阿里云翻译（需要单独配置AccessKey）
	return NewAliyunTranslator(&AliyunTranslateConfig{})
}

// NewWithConfig 使用指定配置创建翻译器
func NewWithConfig(accessKeyID, accessKeySecret string) (Translator, error) {
	config := &AliyunTranslateConfig{
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
	}
	return NewAliyunTranslator(config)
}

// MustNew 创建翻译器，失败时 panic
func MustNew() Translator {
	t, err := New()
	if err != nil {
		panic(err)
	}
	return t
}

// IsConfigured 检查翻译配置是否已设置
func IsConfigured() bool {
	return IsAliyunTranslateConfigured()
}

// GetCurrentConfigSource 返回当前配置来源
func GetCurrentConfigSource() (source, maskedID string) {
	return GetCurrentAliyunTranslateConfigSource()
}

// SaveConfig 保存翻译配置（统一配置接口）
func SaveConfig(accessKeyID, accessKeySecret string) error {
	return SaveAliyunTranslateConfig(accessKeyID, accessKeySecret)
}

// GetAccessKeyID 获取用户配置的 AccessKey ID（统一配置接口）
func GetAccessKeyID() string {
	return GetUserAliyunTranslateAccessKeyID()
}

// GetAccessKeySecret 获取用户配置的 AccessKey Secret（统一配置接口）
func GetAccessKeySecret() string {
	return GetUserAliyunTranslateAccessKeySecret()
}
