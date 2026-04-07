package utils

import (
	"testing"
)

func TestValidateLanguagePair(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		tgt     string
		wantErr bool
	}{
		{"Valid zh to en", "zh", "en", false},
		{"Valid en to zh", "en", "zh", false},
		{"Valid ja to zh", "ja", "zh", false},
		{"Invalid same lang", "zh", "zh", true},
		{"Invalid src lang", "xx", "en", true},
		{"Invalid tgt lang", "zh", "yy", true},
		{"Case insensitive", "ZH", "EN", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLanguagePair(tt.src, tt.tgt)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLanguagePair(%s, %s) error = %v, wantErr %v",
					tt.src, tt.tgt, err, tt.wantErr)
			}
		})
	}
}

func TestIsValidLangCode(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		{"zh", true},
		{"en", true},
		{"ja", true},
		{"ko", true},
		{"fr", true},
		{"de", true},
		{"es", true},
		{"ru", true},
		{"it", true},
		{"pt", true},
		{"XX", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			if got := IsValidLangCode(tt.code); got != tt.want {
				t.Errorf("IsValidLangCode(%s) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestLangCodeToName(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"zh", "中文"},
		{"en", "英文"},
		{"ja", "日文"},
		{"ko", "韩文"},
		{"fr", "法文"},
		{"de", "德文"},
		{"es", "西班牙文"},
		{"ru", "俄文"},
		{"xx", "xx"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			if got := LangCodeToName(tt.code); got != tt.want {
				t.Errorf("LangCodeToName(%s) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}
