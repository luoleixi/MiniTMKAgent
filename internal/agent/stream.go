package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"mini-tmk-agent/internal/audio"
	"mini-tmk-agent/internal/client"
	"mini-tmk-agent/internal/config"
	"mini-tmk-agent/internal/recognizer"
	"mini-tmk-agent/internal/translator"
	"mini-tmk-agent/internal/tts"
	"mini-tmk-agent/internal/utils"
)

// StreamConfig 流式同传配置
type StreamConfig struct {
	SourceLang string // 源语言
	TargetLang string // 目标语言
	DirectMode bool   // 是否使用直连模式（否则使用服务端中转模式）
	ServerURL  string // 服务端地址（服务端模式使用）
}

// StreamAgent 流式同传Agent
type StreamAgent struct {
	config      StreamConfig
	recognizer  recognizer.Recognizer
	recorder    *audio.Recorder
	vad         *audio.VAD
	translator  translator.Translator
	tts         tts.TTS
	player      *audio.Player
	seqGen      *audio.SequenceGenerator
}

// NewStreamAgent 创建流式同传Agent
func NewStreamAgent(config StreamConfig) *StreamAgent {
	return &StreamAgent{
		config: config,
		seqGen: audio.NewSequenceGenerator(),
	}
}

// Start 启动流式翻译
// 默认使用服务端中转模式，如果 DirectMode=true 则使用直连模式
func (a *StreamAgent) Start(ctx context.Context) error {
	if a.config.DirectMode {
		return a.startDirect(ctx)
	}
	return a.startWithServer(ctx)
}

// startWithServer 通过服务端中转运行（默认模式）
func (a *StreamAgent) startWithServer(ctx context.Context) error {
	fmt.Println("🌐 使用服务端中转模式")
	fmt.Println()

	// 创建服务端客户端（使用配置的地址）
	serverURL := a.config.ServerURL
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}
	relayClient := client.NewRelayClient(&client.RelayConfig{
		BaseURL: serverURL,
	})

	// 检查服务端健康状态
	health, err := relayClient.CheckHealth()
	if err != nil {
		return fmt.Errorf("连接服务端失败: %w\n请确保服务端已启动: relay-server start", err)
	}

	fmt.Printf("✅ 服务端连接成功\n")
	if services, ok := health["services"].(map[string]interface{}); ok {
		fmt.Printf("   翻译服务: %v\n", services["translator"])
	}
	fmt.Println()

	// 使用服务端代理服务（ASR、翻译 都通过服务端中转）
	a.translator = relayClient.NewRelayTranslator()
	a.recognizer = relayClient.NewRelayRecognizer(a.config.SourceLang)

	return a.runAudioLoop(ctx)
}

// startDirect 直连百炼平台运行
func (a *StreamAgent) startDirect(ctx context.Context) error {
	// 获取配置
	cfg := config.GetInstance()
	apiKey, ok := cfg.GetBaiwanAPIKey()

	// 检查是否配置了百炼平台 API Key
	if !ok || apiKey == "" {
		fmt.Println()
		fmt.Println("╔════════════════════════════════════════════════════════════╗")
		fmt.Println("║  百炼平台 API Key 未配置                                  ║")
		fmt.Println("╠════════════════════════════════════════════════════════════╣")
		fmt.Println("║                                                            ║")
		fmt.Println("║  直连模式需要配置百炼平台 API Key                         ║")
		fmt.Println("║                                                            ║")
		fmt.Println("║  解决方案：                                                ║")
		fmt.Println("║  1. 使用服务端中转模式（无需 API Key）                    ║")
		fmt.Println("║     ./mini-tmk-agent stream --source-lang zh --target-lang ║")
		fmt.Println("║                                                            ║")
		fmt.Println("║  2. 配置百炼平台 API Key 后使用直连模式                   ║")
		fmt.Println("║     ./mini-tmk-agent config set-baiwan-key <api-key>     ║")
		fmt.Println("║                                                            ║")
		fmt.Println("╚════════════════════════════════════════════════════════════╝")
		fmt.Println()
		return fmt.Errorf("百炼平台 API Key 未配置")
	}

	// 创建识别器（使用百炼平台 WebSocket ASR）
	utils.Debug("创建百炼语音识别器...")
	rec, err := recognizer.NewWithAPIKey(apiKey, a.config.SourceLang)
	if err != nil {
		return fmt.Errorf("创建识别器失败: %w", err)
	}
	a.recognizer = rec

	// 创建翻译器
	trans, err := translator.New()
	if err != nil {
		return fmt.Errorf("创建翻译器失败: %w", err)
	}
	a.translator = trans

	// 创建TTS客户端（直连模式下需要）
	ttsClient, err := tts.New()
	if err != nil {
		fmt.Printf("⚠️ TTS初始化失败: %v\n", err)
		fmt.Println("   语音合成功能将不可用")
	} else {
		a.tts = ttsClient
		fmt.Println("✅ TTS初始化成功")
	}

	return a.runStreamingAudioLoop(ctx)
}

// runAudioLoop 运行音频处理主循环
func (a *StreamAgent) runAudioLoop(ctx context.Context) error {
	// 启动识别器
	if err := a.recognizer.Start(); err != nil {
		return fmt.Errorf("启动识别器失败: %w", err)
	}
	defer a.recognizer.Stop()

	// 显示当前配置信息
	fmt.Printf("🎤 启动流式同传模式 [%s] → [%s]\n", a.config.SourceLang, a.config.TargetLang)
	if a.config.DirectMode {
		asrSource, maskedASRKey := recognizer.GetCurrentAPIKeySource()
		transSource, maskedTransKey := translator.GetCurrentConfigSource()
		fmt.Printf("   模式: 直连阿里云\n")
		fmt.Printf("   ASR 使用 %s: %s\n", asrSource, maskedASRKey)
		fmt.Printf("   翻译使用 %s: %s\n", transSource, maskedTransKey)
	} else {
		serverURL := a.config.ServerURL
		if serverURL == "" {
			serverURL = "http://localhost:8080"
		}
		fmt.Printf("   模式: 服务端中转 (%s)\n", serverURL)
	}
	fmt.Println("   按 Ctrl+C 停止监听...")
	fmt.Println()

	// 创建音频录制器
	recorder, err := audio.NewRecorder(16000)
	if err != nil {
		return fmt.Errorf("创建音频录制器失败: %w", err)
	}
	a.recorder = recorder

	if err := recorder.Start(); err != nil {
		return fmt.Errorf("启动音频录制失败: %w", err)
	}
	defer recorder.Stop()

	// 创建音频播放器
	player, err := audio.NewPlayer(16000)
	if err != nil {
		return fmt.Errorf("创建音频播放器失败: %w", err)
	}
	a.player = player

	if err := player.Start(); err != nil {
		return fmt.Errorf("启动音频播放器失败: %w", err)
	}
	defer player.Stop()

	// 创建 VAD
	vadConfig := audio.DefaultVADConfig()
	vadConfig.MinSpeechDuration = 500
	vadConfig.MinSilenceDuration = 500
	// 降低能量阈值，更容易检测语音
	vadConfig.EnergyThreshold = 0.005

	vad := audio.NewVAD(vadConfig,
		func() {
			// 语音开始
			fmt.Println("🗣️  检测到语音...")
		},
		func(speechData []int16) {
			// 语音结束，发送到识别器
			utils.Debugf("语音结束，数据长度: %d 样本", len(speechData))
			if err := a.recognizer.SendAudio(speechData); err != nil {
				fmt.Printf("   ❌ 发送音频失败: %v\n", err)
			} else {
				utils.Debug("音频已发送到识别器")
			}
		},
	)
	a.vad = vad

	// 启动结果显示协程
	go a.processResults(ctx)

	// 主循环：读取音频并送入 VAD
	fmt.Println("等待语音输入...")
	dataChan := recorder.GetDataChan()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("\n👋 再见!")
			return nil
		case data, ok := <-dataChan:
			if !ok {
				// 通道已关闭，检查是否应该退出
				select {
				case <-ctx.Done():
					fmt.Println("\n👋 再见!")
					return nil
				default:
					// 通道意外关闭，等待一会儿再检查
					time.Sleep(100 * time.Millisecond)
					continue
				}
			}
			// 分帧处理（20ms 一帧）
			frameSize := 320 // 16kHz * 0.02s
			for i := 0; i < len(data); i += frameSize {
				select {
				case <-ctx.Done():
					fmt.Println("\n👋 再见!")
					return nil
				default:
					end := i + frameSize
					if end > len(data) {
						end = len(data)
					}
					frame := data[i:end]
					vad.Process(frame)
				}
			}
		}
	}
}

// processResults 处理识别结果
func (a *StreamAgent) processResults(ctx context.Context) {
	utils.Debug("结果处理协程已启动")
	for {
		select {
		case <-ctx.Done():
			utils.Debug("结果处理协程退出")
			return
		case result, ok := <-a.recognizer.GetResultChan():
			if !ok {
				utils.Debug("结果通道已关闭，退出")
				return
			}
			utils.Debugf("收到识别结果: %s", result.Text)
			if result.Text != "" {
				a.handleRecognitionResult(ctx, result)
			}
		}
	}
}

// handleRecognitionResult 处理单次识别结果
func (a *StreamAgent) handleRecognitionResult(ctx context.Context, result recognizer.RecognitionResult) {
	// 显示识别结果
	if result.IsFinal {
		// 最终结果：换行显示，并进行翻译
		fmt.Printf("\r📝 识别结果: %s\n", result.Text)

		// 进行翻译（仅对最终结果）
		translatedText, err := a.translator.Translate(a.config.SourceLang, a.config.TargetLang, result.Text)
		if err != nil {
			fmt.Printf("   ❌ 翻译失败: %v\n", err)
			return
		}

		fmt.Printf("🌐 翻译结果: %s\n", translatedText)

		// 如果有TTS，进行语音合成（分配序列号确保播放顺序）
		if a.tts != nil {
			seq := a.seqGen.Next()
			go a.playTranslatedText(ctx, seq, translatedText)
		}
	} else {
		// 临时结果：使用回车符覆盖当前行
		fmt.Printf("\r🎤 [识别中] %-50s", result.Text)
	}
}

// playTranslatedText 播放翻译后的文本（异步执行，不阻塞）
// seq: 序列号，确保按句子顺序播放
func (a *StreamAgent) playTranslatedText(ctx context.Context, seq int, text string) {
	// 异步执行 TTS，避免阻塞识别流程
	go func() {
		utils.Debugf("开始TTS合成 [序列号 #%d]...", seq)
		utils.Debug("开始TTS合成...")

		// 检查 TTS 是否可用
		if a.tts == nil {
			fmt.Println("   ❌ TTS 未初始化")
			return
		}

		// 使用超时控制
		ttsDone := make(chan struct{})
		var audioData []byte
		var err error

		go func() {
			audioData, err = a.tts.Synthesize(text, tts.DefaultVoice)
			close(ttsDone)
		}()

		select {
		case <-ttsDone:
			if err != nil {
				fmt.Printf("   ❌ TTS 合成失败: %v\n", err)
				return
			}
		case <-time.After(10 * time.Second):
			fmt.Println("   ⚠️  TTS 合成超时")
			return
		case <-ctx.Done():
			return
		}

		utils.Debugf("TTS 合成成功，音频大小: %.2f KB", float64(len(audioData))/1024)

		fmt.Printf("🔊 播放翻译结果 [序列号 #%d]...\n", seq)
		if err := audio.PlayMP3DataWithSequence(seq, audioData); err != nil {
			fmt.Printf("   ⚠️  播放失败: %v\n", err)
			return
		}

		fmt.Printf("🔊 播放完成 [序列号 #%d]\n", seq)
	}()
}

func (a *StreamAgent) handleRecognizerError(ctx context.Context, err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// 检查是否是连接相关错误或识别器未启动（Token过期后isRunning被设为false）
	isConnectionError := strings.Contains(errStr, "closed network connection") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "websocket") ||
		strings.Contains(errStr, "识别器未启动") ||
		strings.Contains(errStr, "WebSocket 连接未建立")

	if !isConnectionError {
		return false
	}

	fmt.Println("[ASR] 检测到连接断开，尝试自动重连...")

	// 停止旧识别器
	a.recognizer.Stop()

	// 等待一小段时间
	time.Sleep(500 * time.Millisecond)

	// 重新创建识别器
	cfg := config.GetInstance()
	apiKey, _ := cfg.GetBaiwanAPIKey()

	rec, err := recognizer.NewWithAPIKey(apiKey, a.config.SourceLang)
	if err != nil {
		fmt.Printf("[ASR] 创建新识别器失败: %v\n", err)
		return false
	}

	// 启动新识别器
	if err := rec.Start(); err != nil {
		fmt.Printf("[ASR] 启动新识别器失败: %v\n", err)
		return false
	}

	// 替换旧识别器
	a.recognizer = rec

	// 重新启动结果处理协程
	go a.processResults(ctx)

	fmt.Println("[ASR] 自动重连成功，继续识别")
	return true
}

// runStreamingAudioLoop 流式音频循环（实时语音识别 WebSocket 流式传输）
func (a *StreamAgent) runStreamingAudioLoop(ctx context.Context) error {
	// 启动识别器
	if err := a.recognizer.Start(); err != nil {
		return fmt.Errorf("启动识别器失败: %w", err)
	}
	defer a.recognizer.Stop()

	// 创建TTS关闭处理
	if a.tts != nil {
		if closer, ok := a.tts.(interface{ Close() error }); ok {
			defer closer.Close()
		}
	}

	utils.Debug("ASR 识别器已启动")

	// 显示当前配置信息
	fmt.Printf("🎤 启动流式同传模式 [%s] → [%s]\n", a.config.SourceLang, a.config.TargetLang)
	asrSource, maskedASRKey := recognizer.GetCurrentAPIKeySource()
	transSource, maskedTransKey := translator.GetCurrentConfigSource()
	fmt.Printf("   模式: 直连阿里云\n")
	fmt.Printf("   ASR 使用 %s: %s\n", asrSource, maskedASRKey)
	fmt.Printf("   翻译使用 %s: %s\n", transSource, maskedTransKey)
	fmt.Println("   按 Ctrl+C 停止监听...")
	fmt.Println()

	// 创建音频录制器
	recorder, err := audio.NewRecorder(16000)
	if err != nil {
		return fmt.Errorf("创建音频录制器失败: %w", err)
	}
	a.recorder = recorder

	if err := recorder.Start(); err != nil {
		return fmt.Errorf("启动音频录制失败: %w", err)
	}
	utils.Debug("录音已启动")
	defer recorder.Stop()

	// 创建音频播放器
	player, err := audio.NewPlayer(16000)
	if err != nil {
		return fmt.Errorf("创建音频播放器失败: %w", err)
	}
	a.player = player

	if err := player.Start(); err != nil {
		return fmt.Errorf("启动音频播放器失败: %w", err)
	}
	defer player.Stop()

	// 启动结果显示协程
	go a.processResults(ctx)

	// 主循环：直接读取音频并发送到识别器（不使用VAD）
	fmt.Println("等待语音输入...")
	dataChan := recorder.GetDataChan()
	
	// 累积缓冲区，用于批量发送音频
	accumulator := make([]int16, 0, 16000) // 1秒缓冲区
	
	for {
		select {
		case <-ctx.Done():
			fmt.Println("\n👋 再见!")
			return nil
			
		case data, ok := <-dataChan:
			if !ok {
				fmt.Println("\n👋 再见!")
				return nil
			}
			
			// 累积音频数据
			accumulator = append(accumulator, data...)

			// 当累积超过 0.5秒 (8000样本) 时发送
			if len(accumulator) >= 8000 {
				// 发送音频到识别器
				if err := a.recognizer.SendAudio(accumulator); err != nil {
					fmt.Printf("   ⚠️  发送音频失败: %v\n", err)
				} else {
					utils.Debugf("发送音频: %d 样本", len(accumulator))
				}
				// 清空累积器
				accumulator = accumulator[:0]
			}

		default:
			// 如果有累积的数据但不够批量，短暂等待后发送
			if len(accumulator) > 0 {
				select {
				case <-ctx.Done():
					fmt.Println("\n👋 再见!")
					return nil
				case <-time.After(50 * time.Millisecond):
					// 发送剩余音频
					if err := a.recognizer.SendAudio(accumulator); err != nil {
						fmt.Printf("   ⚠️  发送音频失败: %v\n", err)
					}
					accumulator = accumulator[:0]
				}
			} else {
				// 短暂休眠避免忙等待
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}
