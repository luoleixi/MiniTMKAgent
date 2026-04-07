package tts

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// BaiwanTTS 阿里云百炼平台TTS (WebSocket流式API)
type BaiwanTTS struct {
	apiKey     string
	wsURL      string
	sampleRate int
	model      string

	conn     *websocket.Conn
	mu       sync.Mutex
	isClosed bool
}

// NewBaiwanTTS 创建百炼平台TTS客户端
func NewBaiwanTTS(config *Config) (*BaiwanTTS, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("百炼平台API Key不能为空")
	}

	wsURL := config.Endpoint
	if wsURL == "" {
		wsURL = "wss://dashscope.aliyuncs.com/api-ws/v1/inference"
	}

	sampleRate := config.SampleRate
	if sampleRate == 0 {
		sampleRate = DefaultSampleRate
	}

	return &BaiwanTTS{
		apiKey:     config.APIKey,
		wsURL:      wsURL,
		sampleRate: sampleRate,
		model:      DefaultModel,
		isClosed:   true,
	}, nil
}

// Synthesize 将文本合成为语音（非流式，返回完整音频）
func (t *BaiwanTTS) Synthesize(text, voice string) ([]byte, error) {
	if text == "" {
		return nil, fmt.Errorf("合成文本不能为空")
	}

	text = t.cleanText(text)
	if text == "" {
		return nil, fmt.Errorf("清理后文本为空")
	}

	// 建立WebSocket连接
	if err := t.connect(); err != nil {
		return nil, err
	}
	defer t.Close()

	// 收集所有音频数据
	var audioData []byte
	err := t.synthesizeInternal(text, voice, func(chunk []byte) {
		audioData = append(audioData, chunk...)
	})
	if err != nil {
		return nil, err
	}

	return audioData, nil
}

// SynthesizeStream 流式语音合成，通过回调实时接收音频块
func (t *BaiwanTTS) SynthesizeStream(text, voice string, onAudioChunk func(chunk []byte)) error {
	if text == "" {
		return fmt.Errorf("合成文本不能为空")
	}

	text = t.cleanText(text)
	if text == "" {
		return fmt.Errorf("清理后文本为空")
	}

	// 建立WebSocket连接
	if err := t.connect(); err != nil {
		return err
	}

	err := t.synthesizeInternal(text, voice, onAudioChunk)
	// 无论成功与否，都关闭连接，避免状态混乱
	t.Close()
	return err
}

// synthesizeInternal 内部合成方法
// 百炼平台交互流程：run-task -> task-started -> continue-task -> finish-task -> task-finished
func (t *BaiwanTTS) synthesizeInternal(text, voice string, onAudioChunk func(chunk []byte)) error {
	taskID := uuid.New().String()

	if voice == "" {
		voice = DefaultVoice
	}

	// ========== 步骤1: 发送run-task指令 ==========
	runTask := map[string]interface{}{
		"header": map[string]interface{}{
			"action":    "run-task",
			"task_id":   taskID,
			"streaming": "duplex",
		},
		"payload": map[string]interface{}{
			"task_group": "audio",
			"task":       "tts",
			"function":   "SpeechSynthesizer",
			"model":      t.model,
			"parameters": map[string]interface{}{
				"text_type":   "SSML",
				"voice":       voice,
				"format":      "mp3",
				"sample_rate": t.sampleRate,
				"volume":      50,
				"rate":        1,
				"pitch":       1,
			},
			"input": map[string]interface{}{},
		},
	}

	if err := t.conn.WriteJSON(runTask); err != nil {
		return fmt.Errorf("发送run-task失败: %w", err)
	}

	textSent := false
	finished := false
	lastActivity := time.Now()

	// ========== 步骤2-6: 循环接收消息 ==========
	for !finished {
		// 设置读取超时（30秒）
		t.conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		messageType, message, err := t.conn.ReadMessage()
		if err != nil {
			// 检查是否是超时
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				return nil // 正常关闭
			}
			if time.Since(lastActivity) > 30*time.Second {
				return fmt.Errorf("读取WebSocket消息超时")
			}
			return fmt.Errorf("读取WebSocket消息失败: %w", err)
		}
		lastActivity = time.Now()

		// 处理二进制消息（音频数据）
		if messageType == websocket.BinaryMessage {
			if onAudioChunk != nil && len(message) > 0 {
				onAudioChunk(message)
			}
			continue
		}

		// 处理文本消息（JSON事件）
		var event map[string]interface{}
		if err := json.Unmarshal(message, &event); err != nil {
			continue
		}

		header, ok := event["header"].(map[string]interface{})
		if !ok {
			continue
		}

		eventType, _ := header["event"].(string)

		switch eventType {
		// ========== 步骤2: 等待task-started事件 ==========
		case "task-started":
			if !textSent {
				// ========== 步骤3: 发送continue-task指令 ==========
				continueTask := map[string]interface{}{
					"header": map[string]interface{}{
						"action":    "continue-task",
						"task_id":   taskID,
						"streaming": "duplex",
					},
					"payload": map[string]interface{}{
						"input": map[string]interface{}{
							"text": text,
						},
					},
				}

				if err := t.conn.WriteJSON(continueTask); err != nil {
					return fmt.Errorf("发送continue-task失败: %w", err)
				}
				textSent = true

				// 延迟发送finish-task
				time.Sleep(500 * time.Millisecond)

				// ========== 步骤4: 发送finish-task指令 ==========
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

				if err := t.conn.WriteJSON(finishTask); err != nil {
					return fmt.Errorf("发送finish-task失败: %w", err)
				}
			}

		case "result-generated":
			// 服务端正在生成音频，继续接收
			continue

		// ========== 步骤6: 任务结束 ==========
		case "task-finished":
			finished = true
			return nil

		case "task-failed":
			errorCode, _ := header["error_code"].(string)
			errorMessage, _ := header["error_message"].(string)
			return fmt.Errorf("TTS合成任务失败 [%s]: %s", errorCode, errorMessage)
		}
	}

	return nil
}

// connect 建立WebSocket连接
func (t *BaiwanTTS) connect() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// 如果连接已存在，先关闭它
	if t.conn != nil {
		t.conn.Close()
		t.conn = nil
	}
	t.isClosed = true

	headers := http.Header{
		"Authorization":              {fmt.Sprintf("bearer %s", t.apiKey)},
		"X-DashScope-DataInspection": {"enable"},
	}

	conn, _, err := websocket.DefaultDialer.Dial(t.wsURL, headers)
	if err != nil {
		return fmt.Errorf("WebSocket连接失败: %w", err)
	}

	t.conn = conn
	t.isClosed = false
	return nil
}

// cleanText 清理文本（SSML 模式下保留标签格式）
func (t *BaiwanTTS) cleanText(text string) string {
	// 检查是否是 SSML 文本（包含 <speak> 标签）
	isSSML := strings.Contains(text, "<speak") || strings.Contains(text, "<speak>")

	// 只清理 emoji 和特殊控制字符
	replacements := map[string]string{
		"📱": "",
		"🎤": "",
		"🌐": "",
		"🔊": "",
		"⚠️": "",
		"❌": "",
		"✅": "",
		"📝": "",
		"👋": "",
		"🛑": "",
		"⏹️": "",
	}

	for old, new := range replacements {
		text = strings.ReplaceAll(text, old, new)
	}

	// SSML 模式下保留原始格式，不进行空格压缩
	if !isSSML {
		text = strings.TrimSpace(text)
		for strings.Contains(text, "  ") {
			text = strings.ReplaceAll(text, "  ", " ")
		}
	}

	return text
}

// Close 关闭TTS客户端
func (t *BaiwanTTS) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.isClosed = true
	if t.conn != nil {
		conn := t.conn
		t.conn = nil
		return conn.Close()
	}
	return nil
}
