package recognizer

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"mini-tmk-agent/internal/utils"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// BaiwanASR 阿里云百炼平台实时语音识别器（WebSocket直接实现）
type BaiwanASR struct {
	apiKey     string
	wsURL      string
	language   string
	sampleRate int

	conn       *websocket.Conn
	mu         sync.RWMutex
	isRunning  bool
	isClosing  bool
	resultChan chan RecognitionResult
	taskID     string

	// 状态控制
	started   bool
	startedCh chan struct{}
}

// languageMapping 语言代码映射（转换为百炼平台支持的格式）
var languageMapping = map[string]string{
	"zh": "zh",    // 中文
	"en": "en",    // 英文
	"ja": "ja",    // 日语
	"ko": "ko",    // 韩语
	"fr": "fr",    // 法语
	"de": "de",    // 德语
	"es": "es",    // 西班牙语
	"ru": "ru",    // 俄语
	"it": "it",    // 意大利语
}

// normalizeLanguageCode 标准化语言代码
func normalizeLanguageCode(code string) string {
	if mapped, ok := languageMapping[code]; ok {
		return mapped
	}
	return "zh" // 默认中文
}

// newBaiwanASR 创建百炼平台ASR识别器（内部函数）
func newBaiwanASR(apiKey, language string) (*BaiwanASR, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("百炼平台API Key不能为空")
	}

	// 语言映射为百炼支持的格式
	lang := normalizeLanguageCode(language)

	return &BaiwanASR{
		apiKey:     apiKey,
		wsURL:      "wss://dashscope.aliyuncs.com/api-ws/v1/inference",
		language:   lang,
		sampleRate: 16000, // 百炼ASR默认16kHz
		resultChan: make(chan RecognitionResult, 100),
		startedCh:  make(chan struct{}),
	}, nil
}

// Start 启动识别
func (r *BaiwanASR) Start() error {
	r.mu.Lock()
	if r.isRunning {
		r.mu.Unlock()
		return fmt.Errorf("识别器已经在运行")
	}
	r.mu.Unlock()

	// 建立WebSocket连接
	headers := http.Header{
		"Authorization":              {fmt.Sprintf("Bearer %s", r.apiKey)},
		"X-DashScope-DataInspection": {"enable"},
	}

	conn, _, err := websocket.DefaultDialer.Dial(r.wsURL, headers)
	if err != nil {
		return fmt.Errorf("WebSocket连接失败: %w", err)
	}

	// 重置状态
	r.mu.Lock()
	r.conn = conn
	r.isRunning = true
	r.started = false
	r.taskID = uuid.New().String()
	// 重置 startedCh（如果已经关闭，重新创建）
	select {
	case <-r.startedCh:
		r.startedCh = make(chan struct{})
	default:
	}
	r.mu.Unlock()

	// 发送run-task指令
	if err := r.sendRunTask(); err != nil {
		conn.Close()
		r.mu.Lock()
		r.isRunning = false
		r.mu.Unlock()
		return err
	}

	// 启动接收协程
	go r.receiveLoop()

	// 等待 task-started 事件（最多5秒）
	if !r.waitForTaskStarted(5 * time.Second) {
		conn.Close()
		r.mu.Lock()
		r.isRunning = false
		r.mu.Unlock()
		return fmt.Errorf("等待 task-started 事件超时")
	}

	fmt.Println("[ASR] 百炼流式语音识别已就绪")
	return nil
}

// Stop 停止识别
func (r *BaiwanASR) Stop() error {
	r.mu.Lock()
	if !r.isRunning || r.isClosing {
		r.mu.Unlock()
		return nil
	}
	r.isClosing = true
	conn := r.conn
	r.mu.Unlock()

	// 发送finish-task指令
	if conn != nil {
		r.sendFinishTask()
		// 短暂等待
		time.Sleep(100 * time.Millisecond)
		conn.Close()
	}

	r.mu.Lock()
	r.isRunning = false
	r.isClosing = false
	r.started = false
	r.mu.Unlock()

	close(r.resultChan)
	fmt.Println("[ASR] 百炼流式语音识别已断开")
	return nil
}

// SendAudio 发送音频数据
func (r *BaiwanASR) SendAudio(audioData []int16) error {
	r.mu.RLock()
	if !r.isRunning {
		r.mu.RUnlock()
		return fmt.Errorf("识别器未启动")
	}
	if !r.started {
		r.mu.RUnlock()
		return fmt.Errorf("任务尚未开始，请等待task-started事件")
	}
	conn := r.conn
	r.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("WebSocket连接未建立")
	}

	// 将 int16 转为字节（小端序）
	pcmData := make([]byte, len(audioData)*2)
	for i, v := range audioData {
		binary.LittleEndian.PutUint16(pcmData[i*2:], uint16(v))
	}

	// 发送二进制音频数据
	if err := conn.WriteMessage(websocket.BinaryMessage, pcmData); err != nil {
		return fmt.Errorf("发送音频数据失败: %w", err)
	}

	return nil
}

// GetResultChan 获取识别结果通道
func (r *BaiwanASR) GetResultChan() <-chan RecognitionResult {
	return r.resultChan
}

// waitForTaskStarted 等待 task-started 事件
func (r *BaiwanASR) waitForTaskStarted(timeout time.Duration) bool {
	r.mu.RLock()
	if r.started {
		r.mu.RUnlock()
		return true
	}
	r.mu.RUnlock()

	// 等待 channel 或超时
	select {
	case <-r.startedCh:
		return true
	case <-time.After(timeout):
		return false
	}
}

// sendRunTask 发送run-task指令
func (r *BaiwanASR) sendRunTask() error {
	r.mu.RLock()
	taskID := r.taskID
	r.mu.RUnlock()

	runTask := map[string]interface{}{
		"header": map[string]interface{}{
			"action":    "run-task",
			"task_id":   taskID,
			"streaming": "duplex",
		},
		"payload": map[string]interface{}{
			"task_group": "audio",
			"task":       "asr",
			"function":   "recognition",
			"model":      "paraformer-realtime-v2",
			"parameters": map[string]interface{}{
				"language":    r.language,
				"sample_rate": r.sampleRate,
				"format":      "pcm",
			},
			"input": map[string]interface{}{},
		},
	}

	r.mu.RLock()
	conn := r.conn
	r.mu.RUnlock()

	return conn.WriteJSON(runTask)
}

// sendFinishTask 发送finish-task指令
func (r *BaiwanASR) sendFinishTask() error {
	r.mu.RLock()
	taskID := r.taskID
	conn := r.conn
	r.mu.RUnlock()

	if conn == nil {
		return nil
	}

	finishTask := map[string]interface{}{
		"header": map[string]interface{}{
			"action":    "finish-task",
			"task_id":   taskID,
			"streaming": "duplex",
		},
		"payload": map[string]interface{}{
			"input": map[string]interface{}{},
		},
	}

	return conn.WriteJSON(finishTask)
}

// receiveLoop 接收消息循环
func (r *BaiwanASR) receiveLoop() {
	for {
		r.mu.RLock()
		if !r.isRunning {
			r.mu.RUnlock()
			return
		}
		conn := r.conn
		r.mu.RUnlock()

		if conn == nil {
			return
		}

		// 设置读取超时
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		// 读取消息
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				utils.Debugf("WebSocket错误: %v", err)
			}
			return
		}

		// 清除读取超时
		conn.SetReadDeadline(time.Time{})

		// 处理文本消息（JSON事件）
		if messageType == websocket.TextMessage {
			r.handleEvent(message)
		} else if messageType == websocket.BinaryMessage {
			// 二进制消息可能是识别结果
			utils.Debugf("收到二进制消息: %d 字节", len(message))
		}
	}
}

// handleEvent 处理服务端返回的事件
func (r *BaiwanASR) handleEvent(message []byte) {
var event map[string]interface{}
	if err := json.Unmarshal(message, &event); err != nil {
		utils.Debugf("解析事件失败: %v", err)
		return
	}

	header, ok := event["header"].(map[string]interface{})
	if !ok {
		return
	}

	eventType, _ := header["event"].(string)

	switch eventType {
	case "task-started":
		// 任务已启动，可以开始发送音频
		r.mu.Lock()
		r.started = true
		r.mu.Unlock()
		// 通知等待的协程（只关闭一次）
		select {
		case <-r.startedCh:
			// 已经关闭过了
		default:
			close(r.startedCh)
		}
		utils.Info("[ASR] 任务已启动，开始接收音频...")

	case "result-generated":
		// 识别结果
		result := r.parseResult(event)
		if result.Text != "" {
			select {
			case r.resultChan <- result:
			default:
			}
		}

	case "task-finished":
		utils.Debug("任务已完成")

	case "task-failed":
		errorCode, _ := header["error_code"].(string)
		errorMessage, _ := header["error_message"].(string)
		utils.Errorf("任务失败 [%s]: %s", errorCode, errorMessage)

		// 尝试自动刷新Token
		if errorCode == "401" || errorCode == "403" {
			utils.Warn("[ASR] 可能是Token过期，请检查API Key")
		}
	}
}

// parseResult 解析识别结果
func (r *BaiwanASR) parseResult(event map[string]interface{}) RecognitionResult {
	payload, ok := event["payload"].(map[string]interface{})
	if !ok {
		return RecognitionResult{}
	}

	output, ok := payload["output"].(map[string]interface{})
	if !ok {
		return RecognitionResult{}
	}

	sentence, ok := output["sentence"].(map[string]interface{})
	if !ok {
		return RecognitionResult{}
	}

	text, _ := sentence["text"].(string)
	isFinal, _ := sentence["sentence_end"].(bool)

	return RecognitionResult{
		Text:    text,
		IsFinal: isFinal,
	}
}
