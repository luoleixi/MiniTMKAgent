package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mini-tmk-agent/internal/audio"
	"mini-tmk-agent/internal/recognizer"
)

// TranscriptConfig 转录配置
type TranscriptConfig struct {
	AudioFile  string // 音频文件路径
	OutputFile string // 输出文本文件路径
	Language   string // 语言代码
}

// TranscriptAgent 转录Agent
type TranscriptAgent struct {
	config TranscriptConfig
}

// NewTranscriptAgent 创建转录Agent
func NewTranscriptAgent(config TranscriptConfig) *TranscriptAgent {
	if config.Language == "" {
		config.Language = "zh"
	}
	return &TranscriptAgent{
		config: config,
	}
}

// Run 执行转录
func (a *TranscriptAgent) Run(ctx context.Context) error {
	// 检查 API Key
	if !recognizer.IsAPIKeyConfigured() {
		return fmt.Errorf("Kimi API Key 未配置\n请使用 'mini-tmk-agent config set-api-key <your-key>' 设置")
	}

	// 验证音频文件
	if err := a.validateAudioFile(); err != nil {
		return err
	}

	// 显示当前使用的 API Key 来源
	source, maskedKey := recognizer.GetCurrentAPIKeySource()
	fmt.Printf("📝 开始转录音频...\n")
	fmt.Printf("   使用 %s API Key: %s\n", source, maskedKey)
	fmt.Printf("   输入文件: %s\n", a.config.AudioFile)
	fmt.Printf("   语言: %s\n", a.config.Language)
	fmt.Println()

	// 转录音频
	transcript, err := a.transcribeAudio(ctx)
	if err != nil {
		return fmt.Errorf("音频转录失败: %w", err)
	}

	// 写入输出文件
	if err := a.writeOutput(transcript); err != nil {
		return fmt.Errorf("写入输出文件失败: %w", err)
	}

	fmt.Printf("✅ 转录完成，已保存到: %s\n", a.config.OutputFile)
	return nil
}

// validateAudioFile 验证音频文件
func (a *TranscriptAgent) validateAudioFile() error {
	info, err := os.Stat(a.config.AudioFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("音频文件不存在: %s", a.config.AudioFile)
		}
		return err
	}

	if info.IsDir() {
		return fmt.Errorf("指定路径是目录而非文件: %s", a.config.AudioFile)
	}

	// 检查文件扩展名 - 文件转录仅支持MP3格式
	ext := strings.ToLower(filepath.Ext(a.config.AudioFile))
	if ext != ".mp3" {
		return fmt.Errorf("不支持的音频格式: %s, 文件转录仅支持 MP3 格式", ext)
	}

	return nil
}

// transcribeAudio 转录音频 (仅支持MP3)
func (a *TranscriptAgent) transcribeAudio(ctx context.Context) (string, error) {
	fmt.Println("🔄 正在解码MP3音频文件...")

	// 解码音频文件
	samples, err := audio.DecodeFile(a.config.AudioFile)
	if err != nil {
		return "", fmt.Errorf("解码音频失败: %w", err)
	}

	audioDuration := float64(len(samples)) / 16000.0
	fmt.Printf("   音频长度: %.2f 秒\n", audioDuration)
	fmt.Println("🔄 正在识别语音内容...")

	// 创建识别器
	rec, err := recognizer.New(a.config.Language)
	if err != nil {
		return "", fmt.Errorf("创建识别器失败: %w", err)
	}

	if err := rec.Start(); err != nil {
		return "", fmt.Errorf("启动识别器失败: %w", err)
	}

	// 启动结果收集协程
	var fullText strings.Builder
	resultDone := make(chan struct{})
	lastText := "" // 用于去重和保存最后结果

	go func() {
		defer close(resultDone)
		for result := range rec.GetResultChan() {
			if result.Text != "" && result.Text != lastText {
				lastText = result.Text
				if result.IsFinal {
					// 完整句子结果
					fullText.WriteString(result.Text)
					fmt.Printf("   [识别] %s\n", result.Text)
				} else {
					// 中间结果，仅显示不保存
					fmt.Printf("   [识别中] %s\r", result.Text)
				}
			}
		}
		// 如果没有收到最终结果，使用最后一个中间结果
		if fullText.String() == "" && lastText != "" {
			fullText.WriteString(lastText)
		}
	}()

	// 分块发送音频数据（模拟实时流）
	// 使用更大的块大小，提高传输效率
	const chunkSize = 640 // 40ms at 16kHz
	const sendInterval = 40 * time.Millisecond

	ticker := time.NewTicker(sendInterval)
	defer ticker.Stop()

	sampleCount := len(samples)
	sentSamples := 0
	lastProgress := 0

	for sentSamples < sampleCount {
		select {
		case <-ctx.Done():
			rec.Stop()
			return "", ctx.Err()
		case <-ticker.C:
			end := sentSamples + chunkSize
			if end > sampleCount {
				end = sampleCount
			}

			chunk := samples[sentSamples:end]
			if len(chunk) > 0 {
				if err := rec.SendAudio(chunk); err != nil {
					rec.Stop()
					return "", fmt.Errorf("发送音频失败: %w", err)
				}
			}
			sentSamples = end

			// 显示进度（每10%更新一次）
			progress := int(float64(sentSamples) / float64(sampleCount) * 100)
			if progress > lastProgress && progress%10 == 0 {
				fmt.Printf("   进度: %d%%\n", progress)
				lastProgress = progress
			}
		}
	}

	fmt.Println("   进度: 100%")
	fmt.Println("🔄 音频发送完成，等待ASR处理...")

	// 发送1秒静音数据，给ASR时间处理缓冲区的音频
	silence := make([]int16, 16000) // 1秒静音
	for i := 0; i < 5; i++ {      // 分5次发送，每次200ms
		chunk := silence[i*3200 : (i+1)*3200]
		if err := rec.SendAudio(chunk); err != nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Println("🔄 等待识别结果...")
	fmt.Println() // 换行，清除进度行

	// 音频发送完成后，调用 Stop() 发送 finish-task 指令
	// 这会触发服务器返回最终结果并关闭 resultChan
	if err := rec.Stop(); err != nil {
		return "", fmt.Errorf("停止识别器失败: %w", err)
	}

	// 等待结果收集协程完成（带超时）
	select {
	case <-resultDone:
		// 正常完成
	case <-time.After(30 * time.Second):
		return "", fmt.Errorf("等待识别结果超时")
	}

	resultText := fullText.String()
	if resultText == "" {
		resultText = "（未能识别到语音内容）"
	}

	// 构建转录结果
	transcript := fmt.Sprintf(`音频转录结果
==============
文件: %s
时间: %s
语言: %s
音频时长: %.2f 秒

转录内容：
%s
`, a.config.AudioFile, time.Now().Format("2006-01-02 15:04:05"), a.config.Language, audioDuration, resultText)

	return transcript, nil
}

// writeOutput 写入输出文件
func (a *TranscriptAgent) writeOutput(content string) error {
	// 确保输出目录存在
	dir := filepath.Dir(a.config.OutputFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(a.config.OutputFile, []byte(content), 0644)
}
