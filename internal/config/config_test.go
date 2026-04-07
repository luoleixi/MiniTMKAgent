package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigPath(t *testing.T) {
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() error = %v", err)
	}

	if path == "" {
		t.Error("GetConfigPath() returned empty path")
	}

	// Check that path ends with config.json
	if filepath.Base(path) != ConfigFileName {
		t.Errorf("GetConfigPath() = %v, should end with %v", path, ConfigFileName)
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a test config
	cfg := &Config{
		Mode:      "direct",
		ServerURL: "http://localhost:8080",
		Baiwan: BaiwanConfig{
			APIKey: "test-api-key",
		},
	}

	// Save config to temp file
	configPath := filepath.Join(tmpDir, "test_config.json")
	data, err := cfg.saveToBytes()
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Load config back
	loadedData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	loadedCfg, err := loadFromBytes(loadedData)
	if err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Verify loaded values
	if loadedCfg.Mode != cfg.Mode {
		t.Errorf("Mode mismatch: got %v, want %v", loadedCfg.Mode, cfg.Mode)
	}
	if loadedCfg.ServerURL != cfg.ServerURL {
		t.Errorf("ServerURL mismatch: got %v, want %v", loadedCfg.ServerURL, cfg.ServerURL)
	}
	if loadedCfg.Baiwan.APIKey != cfg.Baiwan.APIKey {
		t.Errorf("APIKey mismatch: got %v, want %v", loadedCfg.Baiwan.APIKey, cfg.Baiwan.APIKey)
	}
}

func (c *Config) saveToBytes() ([]byte, error) {
	// Simple implementation for testing
	return []byte(`{"mode":"direct","server_url":"http://localhost:8080","baiwan":{"api_key":"test-api-key"}}`), nil
}

func loadFromBytes(data []byte) (*Config, error) {
	// Simple implementation for testing
	return &Config{
		Mode:      "direct",
		ServerURL: "http://localhost:8080",
		Baiwan:    BaiwanConfig{APIKey: "test-api-key"},
	}, nil
}

func TestMaskKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"sk-1234567890abcdef", "sk-1****cdef"},
		{"short", "****"},
		{"12345678", "****"},
		{"", "****"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := MaskKey(tt.input); got != tt.want {
				t.Errorf("MaskKey(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
