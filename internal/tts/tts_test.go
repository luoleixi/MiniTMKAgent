package tts

import (
	"testing"
)

func TestGetVoiceByLanguage(t *testing.T) {
	tests := []struct {
		lang string
		want string
	}{
		{"zh", "longanyang"},
		{"en", "longanyang"},
		{"ja", "longanyang"},
		{"ko", "longanyang"},
		{"fr", "longanyang"},
		{"de", "longanyang"},
		{"es", "longanyang"},
		{"ru", "longanyang"},
		{"unknown", "longanyang"},
		{"", "longanyang"},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			if got := GetVoiceByLanguage(tt.lang); got != tt.want {
				t.Errorf("GetVoiceByLanguage(%q) = %q, want %q", tt.lang, got, tt.want)
			}
		})
	}
}

func TestLanguageVoiceMap(t *testing.T) {
	// Ensure all supported languages have a voice
	supportedLangs := []string{"zh", "en", "ja", "ko", "fr", "de", "es", "ru"}

	for _, lang := range supportedLangs {
		if _, ok := LanguageVoiceMap[lang]; !ok {
			t.Errorf("LanguageVoiceMap missing entry for %s", lang)
		}
	}
}

func TestBaiwanTTS_cleanText(t *testing.T) {
	tts := &BaiwanTTS{}

	tests := []struct {
		input string
		want  string
	}{
		{"Hello world", "Hello world"},
		{"  Multiple   spaces  ", "Multiple spaces"},
		{"📱 Remove emoji 🎤", "Remove emoji"},
		{"Text with 📱 emoji", "Text with emoji"},
		{"<speak>SSML text</speak>", "<speak>SSML text</speak>"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := tts.cleanText(tt.input)
			if got != tt.want {
				t.Errorf("cleanText(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
