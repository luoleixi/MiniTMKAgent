package recognizer

// RecognitionResult 识别结果
type RecognitionResult struct {
	Text      string  // 识别文本
	IsFinal   bool    // 是否是最终结果
	Confidence float64 // 置信度
}

// Recognizer 语音识别接口
type Recognizer interface {
	// Start 启动识别器
	Start() error
	// Stop 停止识别器
	Stop() error
	// SendAudio 发送音频数据进行识别
	SendAudio(audioData []int16) error
	// GetResultChan 获取识别结果通道
	GetResultChan() <-chan RecognitionResult
}
