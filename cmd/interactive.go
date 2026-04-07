package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"mini-tmk-agent/internal/agent"
	"mini-tmk-agent/internal/config"
	"mini-tmk-agent/internal/utils"
)

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "交互式CLI模式",
	Long:  `启动交互式命令行界面，通过 / 命令进行操作`,
	RunE:  runInteractive,
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
}

func runInteractive(cmd *cobra.Command, args []string) error {
	scanner := bufio.NewScanner(os.Stdin)

	printWelcome()

	// 检查是否已配置
	cfg := config.GetInstance()
	if _, ok := cfg.GetBaiwanAPIKey(); !ok {
		fmt.Println("⚠️  首次使用需要配置百炼 API Key")
		fmt.Println()
		showConfigMenu()
	}

	printPrompt()

	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			printPrompt()
			continue
		}

		// 处理命令
		if strings.HasPrefix(input, "/") {
			if handleCommand(input) {
				break // 退出程序
			}
		} else {
			fmt.Println("  未知输入。输入 /help 查看可用命令")
		}

		printPrompt()
	}

	return nil
}

func printWelcome() {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║                                                          ║")
	fmt.Println("║           MiniTMK Agent - 实时语音同传                   ║")
	fmt.Println("║                                                          ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()
}

func printPrompt() {
	fmt.Println()
	fmt.Println("可用命令:")
	fmt.Println("  /start <源语言> <目标语言> - 启动同传 (如: /start zh en)")
	fmt.Println("  /start-en  - 快速启动 (zh → en)")
	fmt.Println("  /start-ja  - 快速启动 (zh → ja)")
	fmt.Println("  /start-ko  - 快速启动 (zh → ko)")
	fmt.Println("  /transcript <文件> <输出> [语言] - 转录音频文件")
	fmt.Println("  /config    - 查看/修改配置")
	fmt.Println("  /help      - 显示帮助")
	fmt.Println("  /quit      - 退出程序")
	fmt.Println()
}

// parseArgs 解析命令参数，支持引号包裹的带空格路径
func parseArgs(input string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	var quoteChar rune
	prevBackslash := false

	for _, ch := range input {
		if prevBackslash {
			// 前一个字符是反斜杠，直接写入当前字符
			current.WriteRune(ch)
			prevBackslash = false
			continue
		}

		if ch == '\\' {
			// Windows 路径中的反斜杠，检查下一个字符
			if inQuote {
				// 在引号内，直接保留反斜杠和后续字符
				current.WriteRune(ch)
			} else {
				// 不在引号内，可能是转义字符
				current.WriteRune(ch)
				prevBackslash = true
			}
		} else if !inQuote && (ch == '"' || ch == '\'') {
			// 开始引号
			inQuote = true
			quoteChar = ch
			// 如果之前有累积的内容，先保存
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		} else if inQuote && ch == quoteChar {
			// 结束引号
			inQuote = false
			quoteChar = 0
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		} else if !inQuote && (ch == ' ' || ch == '\t') {
			// 空格分隔符
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		} else {
			// 普通字符
			current.WriteRune(ch)
		}
	}

	// 添加最后一个参数
	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

func handleCommand(input string) bool {
	parts := parseArgs(input)
	if len(parts) == 0 {
		return false
	}

	command := parts[0]

	switch command {
	case "/quit", "/exit", "/q":
		fmt.Println("\n👋 再见!")
		return true

	case "/help", "/h", "?":
		printHelp()

	case "/start":
		// 支持多种格式: /start, /start zh en, /start en zh
		var src, tgt string
		switch len(parts) {
		case 1:
			// /start 单独使用，使用默认值
			src, tgt = "zh", "en"
		case 2:
			// /start en -> 假设源语言是 zh
			src, tgt = "zh", parts[1]
		case 3:
			// /start zh en -> 完整指定
			src, tgt = parts[1], parts[2]
		default:
			fmt.Println("  用法: /start [源语言] [目标语言]")
			fmt.Println("  示例: /start, /start en, /start zh en")
			return false
		}
		if err := utils.ValidateLanguagePair(src, tgt); err != nil {
			fmt.Printf("  ❌ %v\n", err)
			return false
		}
		startStreaming(src, tgt)

	case "/start-en":
		startStreaming("zh", "en")

	case "/start-ja":
		startStreaming("zh", "ja")

	case "/start-ko":
		startStreaming("zh", "ko")

	case "/start-fr":
		startStreaming("zh", "fr")

	case "/start-de":
		startStreaming("zh", "de")

	case "/start-es":
		startStreaming("zh", "es")

	case "/start-ru":
		startStreaming("zh", "ru")

	case "/transcript", "/trans":
		// /transcript <文件> <输出> [语言]
		if len(parts) < 3 {
			fmt.Println("  用法: /transcript <音频文件> <输出文件> [语言]")
			fmt.Println("  示例: /transcript audio.wav output.txt zh")
			fmt.Println("  路径含空格请用引号: /transcript \"D:\\path\\file.wav\" out.txt zh")
			return false
		}
		audioFile := parts[1]
		outputFile := parts[2]
		lang := "zh"
		if len(parts) >= 4 {
			lang = parts[3]
		}
		if err := utils.ValidateLanguagePair(lang, "en"); err != nil {
			// 只需要验证语言代码是否有效，目标语言无所谓
			if !utils.IsValidLangCode(lang) {
				fmt.Printf("  ❌ 无效的语言代码: %s\n", lang)
				return false
			}
		}
		runTranscriptInteractive(audioFile, outputFile, lang)

	case "/config", "/settings":
		showConfigMenu()

	default:
		fmt.Printf("  未知命令: %s\n", command)
		fmt.Println("  输入 /help 查看可用命令")
	}

	return false
}

func printHelp() {
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("                     命令帮助")
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("【启动同传】")
	fmt.Printf("  %-15s 启动同传 (中文 → 英文)\n", "/start")
	fmt.Printf("  %-15s 启动同传: /start [源语言] [目标语言]\n", "/start <src> <tgt>")
	fmt.Println("                   示例: /start zh en, /start en zh")
	fmt.Println()
	fmt.Println("【快速启动】")
	fmt.Printf("  %-15s 快速启动 (中文 → 英文)\n", "/start-en")
	fmt.Printf("  %-15s 快速启动 (中文 → 日文)\n", "/start-ja")
	fmt.Printf("  %-15s 快速启动 (中文 → 韩文)\n", "/start-ko")
	fmt.Printf("  %-15s 快速启动 (中文 → 法文)\n", "/start-fr")
	fmt.Printf("  %-15s 快速启动 (中文 → 德文)\n", "/start-de")
	fmt.Printf("  %-15s 快速启动 (中文 → 西班牙文)\n", "/start-es")
	fmt.Printf("  %-15s 快速启动 (中文 → 俄文)\n", "/start-ru")
	fmt.Println()
	fmt.Println("【支持的语言代码】")
	fmt.Printf("  zh = %s  ", utils.LangCodeToName("zh"))
	fmt.Printf("en = %s  ", utils.LangCodeToName("en"))
	fmt.Printf("ja = %s  ", utils.LangCodeToName("ja"))
	fmt.Printf("ko = %s\n", utils.LangCodeToName("ko"))
	fmt.Printf("  fr = %s  ", utils.LangCodeToName("fr"))
	fmt.Printf("de = %s  ", utils.LangCodeToName("de"))
	fmt.Printf("es = %s  ", utils.LangCodeToName("es"))
	fmt.Printf("ru = %s\n", utils.LangCodeToName("ru"))
	fmt.Println()
	fmt.Println("【音频转录】")
	fmt.Printf("  %-15s 转录音频文件\n", "/transcript")
	fmt.Println("                   用法: /transcript <音频文件> <输出文件> [语言]")
	fmt.Println("                   示例: /transcript audio.wav out.txt zh")
	fmt.Println("                   支持格式: wav, mp3, pcm, m4a, flac, aac, ogg")
	fmt.Println()
	fmt.Println("【配置管理】")
	fmt.Printf("  %-15s 查看和修改配置\n", "/config")
	fmt.Println()
	fmt.Println("【其他】")
	fmt.Printf("  %-15s 显示此帮助\n", "/help")
	fmt.Printf("  %-15s 退出程序\n", "/quit")
	fmt.Println()
}

func runTranscriptInteractive(audioFile, outputFile, lang string) {
	fmt.Printf("\n📝 开始转录 [%s] -> [%s]\n", audioFile, outputFile)
	fmt.Printf("   语言: %s\n\n", lang)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建转录Agent
	transcriptAgent := agent.NewTranscriptAgent(agent.TranscriptConfig{
		AudioFile:  audioFile,
		OutputFile: outputFile,
		Language:   lang,
	})

	// 执行转录
	if err := transcriptAgent.Run(ctx); err != nil {
		fmt.Printf("❌ 转录失败: %v\n", err)
	}

	fmt.Println()
}

func startStreaming(sourceLang, targetLang string) {
	fmt.Printf("\n🎤 启动同传 [%s] → [%s]\n", sourceLang, targetLang)
	fmt.Println("   按 Ctrl+C 停止同传，返回主菜单")
	fmt.Println()

	// 创建上下文，支持信号中断
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 捕获中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n🛑 停止同传...")
		cancel()
	}()

	// 从配置读取模式
	cfg := config.GetInstance()
	mode, serverURL := cfg.GetMode()
	isDirectMode := mode == "direct"

	// 创建流式Agent
	streamAgent := agent.NewStreamAgent(agent.StreamConfig{
		SourceLang: sourceLang,
		TargetLang: targetLang,
		DirectMode: isDirectMode,
		ServerURL:  serverURL,
	})

	// 启动流式翻译
	if err := streamAgent.Start(ctx); err != nil {
		fmt.Printf("❌ 启动失败: %v\n", err)
	}

	fmt.Println()
}

func showConfigMenu() {
	cfg := config.GetInstance()
	apiKey, _ := cfg.GetBaiwanAPIKey()
	mode, serverURL := cfg.GetMode()

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("                     当前配置")
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println()

	if apiKey != "" {
		maskedKey := apiKey
		if len(apiKey) > 8 {
			maskedKey = apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
		}
		fmt.Printf("  API Key:   %s\n", maskedKey)
	} else {
		fmt.Println("  API Key:   未配置")
	}

	fmt.Printf("  模式:      %s\n", mode)
	if mode == "relay" {
		fmt.Printf("  服务器:    %s\n", serverURL)
	}

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("配置选项:")
	fmt.Println("  1. 设置 API Key")
	fmt.Println("  2. 切换模式 (直连/中继)")
	fmt.Println("  3. 返回主菜单")
	fmt.Println()
	fmt.Print("请选择 [1-3]: ")

	var choice string
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		fmt.Print("\n请输入百炼 API Key: ")
		var newKey string
		fmt.Scanln(&newKey)
		if newKey != "" {
			if err := cfg.SetBaiwanAPIKey(newKey); err != nil {
				fmt.Printf("❌ 保存失败: %v\n", err)
			} else {
				fmt.Println("✅ API Key 已保存")
			}
		}

	case "2":
		fmt.Println("\n选择模式:")
		fmt.Println("  1. 直连模式 (使用本地 API Key)")
		fmt.Println("  2. 中继模式 (通过服务器中转)")
		fmt.Print("\n请选择 [1-2]: ")

		var modeChoice string
		fmt.Scanln(&modeChoice)

		switch modeChoice {
		case "1":
			cfg.SetMode("direct", "")
			fmt.Println("✅ 已切换到直连模式")
		case "2":
			fmt.Print("请输入服务器地址 (默认: http://localhost:8080): ")
			var server string
			fmt.Scanln(&server)
			if server == "" {
				server = "http://localhost:8080"
			}
			cfg.SetMode("relay", server)
			fmt.Println("✅ 已切换到中继模式")
		}

	default:
		fmt.Println("返回主菜单")
	}

	fmt.Println()
}
