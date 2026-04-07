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
