package setup

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"mini-tmk-agent/internal/config"
)

// RunInteractiveSetup 运行交互式配置引导
func RunInteractiveSetup() error {
	fmt.Println("========================================")
	fmt.Println("  🎤 MiniTMK Agent 首次配置向导")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("欢迎使用 MiniTMK Agent！")
	fmt.Println("本程序支持两种运行模式：")
	fmt.Println("  1. 服务端中转模式 - 通过服务端处理请求（推荐）")
	fmt.Println("  2. 直连模式 - 本地直接调用百炼平台 API")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// 步骤 1: 选择运行模式
	mode, serverURL, err := selectMode(reader)
	if err != nil {
		return err
	}

	// 步骤 2: 根据模式进行配置
	if mode == "server" {
		err = configureServerMode(reader, serverURL)
	} else {
		err = configureDirectMode(reader)
	}

	if err != nil {
		return err
	}

	// 保存模式配置
	cfg := config.GetInstance()
	if err := cfg.SetMode(mode, serverURL); err != nil {
		fmt.Printf("⚠️  警告: 保存模式配置失败: %v\n", err)
	}

	fmt.Println()
	fmt.Println("✅ 配置完成！")
	if mode == "server" {
		fmt.Printf("   模式: 服务端中转 (%s)\n", serverURL)
	} else {
		fmt.Println("   模式: 直连百炼平台")
	}
	fmt.Println()
	fmt.Println("现在可以运行:")
	if mode == "server" {
		fmt.Println("   mini-tmk-agent stream --source-lang zh --target-lang en")
	} else {
		fmt.Println("   mini-tmk-agent stream --source-lang zh --target-lang en --direct")
	}
	fmt.Println()

	return nil
}

// selectMode 选择运行模式
func selectMode(reader *bufio.Reader) (mode, serverURL string, err error) {
	for {
		fmt.Println("请选择运行模式:")
		fmt.Println("  1) 服务端中转模式（默认: localhost:8080）")
		fmt.Println("  2) 直连百炼平台模式（需要配置 API Key）")
		fmt.Print("请输入选项 [1/2] (默认 1): ")

		input, err := readInput(reader)
		if err != nil {
			fmt.Printf("⚠️  读取输入失败: %v，请重试\n\n", err)
			continue
		}

		// 默认选择服务端模式
		if input == "" {
			input = "1"
		}

		switch input {
		case "1", "server", "s":
			// 服务端模式，询问服务器地址
			serverURL, err = inputServerURL(reader)
			if err != nil {
				fmt.Printf("⚠️  %v，请重试\n\n", err)
				continue
			}
			return "server", serverURL, nil

		case "2", "direct", "d":
			// 直连模式
			return "direct", "", nil

		default:
			fmt.Printf("⚠️  无效选项 '%s'，请输入 1 或 2\n\n", input)
			continue
		}
	}
}

// inputServerURL 输入服务器地址
func inputServerURL(reader *bufio.Reader) (string, error) {
	for {
		fmt.Print("请输入服务端地址 [默认: http://localhost:8080]: ")
		input, err := readInput(reader)
		if err != nil {
			return "", fmt.Errorf("读取输入失败: %w", err)
		}

		// 使用默认值
		if input == "" {
			return "http://localhost:8080", nil
		}

		// 简单验证 URL 格式
		if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
			fmt.Println("⚠️  地址必须以 http:// 或 https:// 开头，请重试")
			continue
		}

		return input, nil
	}
}

// configureServerMode 配置服务端模式
func configureServerMode(reader *bufio.Reader, serverURL string) error {
	fmt.Println()
	fmt.Println("🌐 服务端中转模式配置")
	fmt.Println("--------------------")
	fmt.Printf("服务端地址: %s\n", serverURL)
	fmt.Println()
	fmt.Println("此模式下 ASR/翻译/TTS 都通过服务端处理，")
	fmt.Println("本地无需配置 API Key。")
	fmt.Println()
	return nil
}

// configureDirectMode 配置直连模式
func configureDirectMode(reader *bufio.Reader) error {
	fmt.Println()
	fmt.Println("📡 直连百炼平台模式配置")
	fmt.Println("----------------------")
	fmt.Println()
	fmt.Println("📖 获取百炼平台 API Key：")
	fmt.Println("   1. 访问 https://dashscope.console.aliyun.com/")
	fmt.Println("   2. 进入 'API Key管理' 创建 API Key")
	fmt.Println("   3. 复制 API Key")
	fmt.Println()

	// 配置百炼平台 API Key
	return configureBaiwanAPIKey(reader)
}

// configureBaiwanAPIKey 配置百炼平台 API Key
func configureBaiwanAPIKey(reader *bufio.Reader) error {
	fmt.Println("【百炼平台 API Key 配置】")
	fmt.Println()

	// API Key
	apiKey, err := promptRequired(reader, "请输入百炼平台 API Key: ")
	if err != nil {
		return err
	}

	// 保存
	fmt.Println("正在保存百炼平台配置...")
	cfg := config.GetInstance()
	if err := cfg.SetBaiwanAPIKey(apiKey); err != nil {
		return fmt.Errorf("保存百炼平台配置失败: %w", err)
	}

	fmt.Printf("✅ 百炼平台配置已保存: %s\n", maskKey(apiKey))
	fmt.Println()
	fmt.Println("现在可以使用:")
	fmt.Println("  ✓ 语音识别 (ASR)")
	fmt.Println("  ✓ 机器翻译 (Translate)")
	fmt.Println("  ✓ 语音合成 (TTS)")
	fmt.Println()
	return nil
}

// promptRequired 提示输入必填项（带错误处理和重试）
func promptRequired(reader *bufio.Reader, prompt string) (string, error) {
	for {
		fmt.Print(prompt)
		input, err := readInput(reader)
		if err != nil {
			fmt.Printf("⚠️  读取输入失败: %v，请重试\n", err)
			continue
		}

		if input == "" {
			fmt.Println("⚠️  此项不能为空，请重新输入")
			continue
		}

		return input, nil
	}
}

// readInput 读取用户输入
func readInput(reader *bufio.Reader) (string, error) {
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// CheckAndSetup 检查配置，如果没有则运行交互式配置
func CheckAndSetup() error {
	cfg := config.GetInstance()

	// 检查是否已设置运行模式
	mode, _ := cfg.GetMode()

	// 如果已设置模式为 server，直接返回
	if mode == "server" {
		return nil
	}

	// 如果已设置模式为 direct，检查是否配置了百炼 API Key
	if mode == "direct" {
		if cfg.IsConfigured() {
			return nil
		}
	}

	// 未配置，运行交互式引导
	return RunInteractiveSetup()
}

// Confirm 提示用户确认
func Confirm(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", prompt)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// maskKey 安全地掩码显示密钥
func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}
