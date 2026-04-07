package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	// ConfigDirName 配置目录名
	ConfigDirName = "mini-tmk-agent"
	// ConfigFileName 配置文件名
	ConfigFileName = "config.json"
)

// Config 应用配置
// 简化版本：统一使用百炼平台 API Key，无需配置阿里云
type Config struct {
	Mode      string       `json:"mode"`       // "server" 或 "direct"
	ServerURL string       `json:"server_url"` // 服务端地址
	Baiwan    BaiwanConfig `json:"baiwan"`     // 百炼平台配置（统一用于 ASR 和 TTS）
}

// BaiwanConfig 百炼平台配置
type BaiwanConfig struct {
	APIKey string `json:"api_key"`
}

var (
	instance *Config
	once     sync.Once
	mu       sync.RWMutex
)

// GetInstance 获取配置单例
func GetInstance() *Config {
	once.Do(func() {
		instance = loadConfig()
	})
	return instance
}

// loadConfig 从文件加载配置
func loadConfig() *Config {
	config := &Config{}

	configPath, err := GetConfigPath()
	if err != nil {
		return config
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return config
	}

	if err := json.Unmarshal(data, config); err != nil {
		return &Config{}
	}

	return config
}

// Save 保存配置到文件
func (c *Config) Save() error {
	mu.Lock()
	defer mu.Unlock()

	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("获取配置目录失败: %w", err)
	}

	appDir := filepath.Join(configDir, ConfigDirName)
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	configPath := filepath.Join(appDir, ConfigFileName)
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// Windows 使用 0666 权限（兼容性好）
	perm := os.FileMode(0600)
	if os.PathSeparator == '\\' {
		perm = 0666
	}

	if err := os.WriteFile(configPath, data, perm); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// GetConfigPath 获取配置文件路径
func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, ConfigDirName, ConfigFileName), nil
}

// SetBaiwanAPIKey 设置百炼平台API Key
func (c *Config) SetBaiwanAPIKey(apiKey string) error {
	mu.Lock()
	c.Baiwan.APIKey = apiKey
	mu.Unlock()

	return c.Save()
}

// GetBaiwanAPIKey 获取百炼平台API Key
func (c *Config) GetBaiwanAPIKey() (apiKey string, ok bool) {
	mu.RLock()
	defer mu.RUnlock()
	return c.Baiwan.APIKey, c.Baiwan.APIKey != ""
}

// IsConfigured 检查是否已配置百炼平台 API Key
func (c *Config) IsConfigured() bool {
	mu.RLock()
	defer mu.RUnlock()
	return c.Baiwan.APIKey != ""
}

// SetMode 设置运行模式
func (c *Config) SetMode(mode, serverURL string) error {
	mu.Lock()
	c.Mode = mode
	c.ServerURL = serverURL
	mu.Unlock()

	return c.Save()
}

// GetMode 获取运行模式
func (c *Config) GetMode() (mode, serverURL string) {
	mu.RLock()
	defer mu.RUnlock()

	if c.Mode == "" {
		return "server", c.ServerURL
	}
	return c.Mode, c.ServerURL
}

// Clear 清除所有配置
func (c *Config) Clear() error {
	mu.Lock()
	c.Mode = ""
	c.ServerURL = ""
	c.Baiwan = BaiwanConfig{}
	mu.Unlock()

	return c.Save()
}

// MaskKey 隐藏密钥中间部分
func MaskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}
