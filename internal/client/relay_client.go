package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"mini-tmk-agent/internal/recognizer"
	"mini-tmk-agent/internal/translator"
	"mini-tmk-agent/internal/tts"
	"mini-tmk-agent/internal/utils"
)

// RelayClient 中继服务器客户端
type RelayClient struct {
	baseURL string
	client  *http.Client
}

// RelayConfig 中继配置
type RelayConfig struct {
	BaseURL string // 中继服务器地址，如 "http://localhost:8080"
}

// NewRelayClient 创建中继客户端
func NewRelayClient(config *RelayConfig) *RelayClient {
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:8080"
	}
	return &RelayClient{
		baseURL: config.BaseURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// RelayTranslator 使用中继服务的翻译器
type RelayTranslator struct {
	client *RelayClient
}

// NewRelayTranslator 创建中继翻译器
func (c *RelayClient) NewRelayTranslator() translator.Translator {
	return &RelayTranslator{client: c}
}

// Translate 通过中继服务器翻译
func (t *RelayTranslator) Translate(sourceLang, targetLang, text string) (string, error) {
	reqBody, _ := json.Marshal(map[string]string{
		"source_lang": sourceLang,
		"target_lang": targetLang,
		"text":        text,
	})

	resp, err := t.client.client.Post(
		t.client.baseURL+"/api/translate",
		"application/json",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return "", fmt.Errorf("请求中继服务器失败: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Success    bool   `json:"success"`
		Translated string `json:"translated"`
		Error      string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if !result.Success {
		return "", fmt.Errorf("翻译失败: %s", result.Error)
	}

	return result.Translated, nil
}

// RelayTTS 使用中继服务的 TTS
type RelayTTS struct {
	client *RelayClient
}

// NewRelayTTS 创建中继 TTS
func (c *RelayClient) NewRelayTTS() tts.TTS {
	return &RelayTTS{client: c}
}

// Synthesize 通过中继服务器合成语音
func (t *RelayTTS) Synthesize(text, voice string) ([]byte, error) {
	reqBody, _ := json.Marshal(map[string]string{
		"text":  text,
		"voice": voice,
		"lang":  "zh",
	})

	resp, err := t.client.client.Post(
		t.client.baseURL+"/api/tts",
		"application/json",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("请求中继服务器失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TTS 失败: %s", string(body))
	}

	return io.ReadAll(resp.Body)
}

// SynthesizeStream 流式合成（中继模式不支持真正的流式，返回完整数据）
func (t *RelayTTS) SynthesizeStream(text, voice string, onAudioChunk func(chunk []byte)) error {
	audioData, err := t.Synthesize(text, voice)
	if err != nil {
		return err
	}

	// 回调音频数据
	if onAudioChunk != nil {
		onAudioChunk(audioData)
	}

	return nil
}

// Close 关闭客户端
func (t *RelayTTS) Close() error {
	return nil
}

// RelayRecognizer 使用中继服务的识别器
type RelayRecognizer struct {
	client     *RelayClient
	resultChan chan recognizer.RecognitionResult
	language   string
}

// NewRelayRecognizer 创建中继识别器
func (c *RelayClient) NewRelayRecognizer(language string) recognizer.Recognizer {
	return &RelayRecognizer{
		client:     c,
		resultChan: make(chan recognizer.RecognitionResult, 10),
		language:   language,
	}
}

// Start 启动识别器
func (r *RelayRecognizer) Start() error {
	return nil
}

// Stop 停止识别器
func (r *RelayRecognizer) Stop() error {
	close(r.resultChan)
	return nil
}

// SendAudio 发送音频数据到服务端进行识别（异步非阻塞）
func (r *RelayRecognizer) SendAudio(audioData []int16) error {
	// 异步处理，避免阻塞音频采集
	go r.recognizeAsync(audioData)
	return nil
}

// recognizeAsync 异步进行语音识别
func (r *RelayRecognizer) recognizeAsync(audioData []int16) {
	// 将 int16 音频数据转换为字节
	pcmData := make([]byte, len(audioData)*2)
	for i, v := range audioData {
		pcmData[i*2] = byte(v)
		pcmData[i*2+1] = byte(v >> 8)
	}

	utils.Debugf("发送 ASR 请求，音频大小: %d 字节", len(pcmData))

	// 调用服务端 ASR API
	url := fmt.Sprintf("%s/api/asr?lang=%s", r.client.baseURL, r.language)
	resp, err := r.client.client.Post(url, "application/octet-stream", bytes.NewReader(pcmData))
	if err != nil {
		utils.Debugf("ASR 请求失败: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Debugf("读取响应失败: %v", err)
		return
	}

	utils.Debugf("ASR 响应: %s", string(body))

	var result struct {
		Success bool   `json:"success"`
		Text    string `json:"text"`
		Error   string `json:"error"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		utils.Debugf("解析响应失败: %v", err)
		return
	}

	if !result.Success {
		utils.Debugf("ASR 识别失败: %s", result.Error)
		return
	}

	// 发送识别结果到通道
	if result.Text != "" {
		utils.Debugf("ASR 识别成功: %s", result.Text)
		r.resultChan <- recognizer.RecognitionResult{
			Text:       result.Text,
			IsFinal:    true,
			Confidence: 0.9,
		}
	}
}

// GetResultChan 获取识别结果通道
func (r *RelayRecognizer) GetResultChan() <-chan recognizer.RecognitionResult {
	return r.resultChan
}

// CheckHealth 检查中继服务器健康状态
func (c *RelayClient) CheckHealth() (map[string]interface{}, error) {
	resp, err := c.client.Get(c.baseURL + "/api/health")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// ASRTokenResponse ASR Token 响应
type ASRTokenResponse struct {
	Success  bool   `json:"success"`
	AppKey   string `json:"app_key,omitempty"`
	Token    string `json:"token,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
	Expire   int64  `json:"expire,omitempty"`
	Error    string `json:"error,omitempty"`
}

// GetASRToken 从 relay-server 获取 ASR Token（用于客户端直连阿里云）
func (c *RelayClient) GetASRToken() (*ASRTokenResponse, error) {
	resp, err := c.client.Get(c.baseURL + "/api/token/asr")
	if err != nil {
		return nil, fmt.Errorf("请求 ASR Token 失败: %w", err)
	}
	defer resp.Body.Close()

	var result ASRTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("获取 ASR Token 失败: %s", result.Error)
	}

	return &result, nil
}
