package tests

import (
	"testing"

	"mini-tmk-agent/internal/audio"
	"mini-tmk-agent/internal/config"
	"mini-tmk-agent/internal/utils"
)

// TestEndToEndFlow 测试完整流程
func TestEndToEndFlow(t *testing.T) {
	t.Run("Language Validation Flow", func(t *testing.T) {
		// 测试语言验证
		langs := []struct {
			src  string
			tgt  string
			want bool
		}{
			{"zh", "en", true},
			{"en", "ja", true},
			{"zh", "zh", false},
		}

		for _, l := range langs {
			err := utils.ValidateLanguagePair(l.src, l.tgt)
			if (err == nil) != l.want {
				t.Errorf("ValidateLanguagePair(%s, %s) error = %v, want valid = %v",
					l.src, l.tgt, err, l.want)
			}
		}
	})

	t.Run("Config Operations", func(t *testing.T) {
		// 测试配置路径获取
		path, err := config.GetConfigPath()
		if err != nil {
			t.Skipf("Config not available: %v", err)
		}
		if path == "" {
			t.Error("GetConfigPath returned empty string")
		}

		// 测试 API Key 掩码
		masked := config.MaskKey("sk-1234567890abcdef")
		if masked == "sk-1234567890abcdef" {
			t.Error("MaskKey did not mask the key")
		}
	})

	t.Run("PlayQueue Sequence", func(t *testing.T) {
		queue := audio.GetPlayQueue()

		// 测试序列号生成
		seq1 := queue.NextSequence()
		seq2 := queue.NextSequence()

		if seq2 != seq1+1 {
			t.Errorf("Sequence not incrementing: got %d after %d", seq2, seq1)
		}

		// 重置
		queue.Reset()
		if queue.NextSequence() != 1 {
			t.Error("Reset did not reset sequence")
		}
	})
}
