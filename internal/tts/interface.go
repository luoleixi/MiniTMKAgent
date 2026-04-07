package tts

// TTS 语音合成接口
type TTS interface {
	// Synthesize 将文本合成为语音，返回音频数据(MP3格式)
	// text: 要合成的文本
	// voice: 音色名称
	Synthesize(text, voice string) ([]byte, error)

	// SynthesizeStream 流式语音合成，实时播放
	// text: 要合成的文本
	// voice: 音色名称
	// onAudioChunk: 音频块回调函数，用于实时播放
	SynthesizeStream(text, voice string, onAudioChunk func(chunk []byte)) error

	// Close 关闭 TTS 客户端
	Close() error
}

// Config TTS 配置
type Config struct {
	APIKey     string
	Endpoint   string
	SampleRate int
}

// DefaultSampleRate 默认采样率
const DefaultSampleRate = 22050

// DefaultVoice 默认音色
const DefaultVoice = "longanyang"

// DefaultModel 默认模型
const DefaultModel = "cosyvoice-v3-flash"

// LanguageVoiceMap 语言到音色的映射
// CosyVoice-v3 支持多语言跨语言合成，使用通用音色即可
var LanguageVoiceMap = map[string]string{
	"zh": "longanyang", // 中文（多语言通用）
	"en": "longanyang", // 英文（多语言通用）
	"ja": "longanyang", // 日文（多语言通用）
	"ko": "longanyang", // 韩文（多语言通用）
	"fr": "longanyang", // 法文（多语言通用）
	"de": "longanyang", // 德文（多语言通用）
	"es": "longanyang", // 西班牙文（多语言通用）
	"ru": "longanyang", // 俄文（多语言通用）
}

// GetVoiceByLanguage 根据语言代码获取音色
func GetVoiceByLanguage(lang string) string {
	if voice, ok := LanguageVoiceMap[lang]; ok {
		return voice
	}
	return DefaultVoice
}
