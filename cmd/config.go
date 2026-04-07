package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"mini-tmk-agent/internal/config"
	"mini-tmk-agent/internal/tts"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置管理",
	Long:  "管理 mini-tmk-agent 的配置（统一使用百炼平台 API Key）",
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看配置状态",
	Long:  "显示当前配置的状态",
	RunE:  runStatus,
}

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "清除所有配置",
	Long:  "删除所有用户配置",
	RunE:  runClear,
}

var setBaiwanCmd = &cobra.Command{
	Use:   "set-baiwan-key",
	Short: "设置百炼平台API Key",
	Long: `设置阿里云百炼平台的API Key，用于语音识别(ASR)和语音合成(TTS)。

获取方式:
  1. 登录阿里云控制台: https://dashscope.console.aliyun.com/
  2. 进入"API Key管理"创建API Key
  3. 复制API Key并设置

注意: 百炼平台 API Key 同时用于 ASR 和 TTS，只需配置一次即可`,
	Example: `mini-tmk-agent config set-baiwan-key sk-xxxxxx`,
	RunE:    runSetBaiwanKey,
}

var setModeCmd = &cobra.Command{
	Use:   "set-mode",
	Short: "设置运行模式",
	Long: `设置运行模式：服务端中转模式 或 直连模式。

模式说明:
  - server: 服务端中转模式，通过服务端处理请求（推荐）
  - direct: 直连模式，本地直接调用百炼平台 API（需配置百炼 API Key）

示例:
  mini-tmk-agent config set-mode server --server-url http://localhost:8080
  mini-tmk-agent config set-mode direct`,
	RunE: runSetMode,
}

var getModeCmd = &cobra.Command{
	Use:   "get-mode",
	Short: "查看当前运行模式",
	Long:  "显示当前配置的运行模式",
	RunE:  runGetMode,
}

var (
	baiwanAPIKey    string
	serverURLFlag   string
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(statusCmd)
	configCmd.AddCommand(clearCmd)
	configCmd.AddCommand(setBaiwanCmd)
	configCmd.AddCommand(setModeCmd)
	configCmd.AddCommand(getModeCmd)

	// set-baiwan-key 命令参数
	setBaiwanCmd.Flags().StringVarP(&baiwanAPIKey, "api-key", "k", "", "百炼平台API Key")
	setBaiwanCmd.MarkFlagRequired("api-key")

	// set-mode 命令参数
	setModeCmd.Flags().StringVarP(&serverURLFlag, "server-url", "u", "http://localhost:8080", "服务端地址（server 模式使用）")
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg := config.GetInstance()

	fmt.Println("配置状态")
	fmt.Println("========")
	fmt.Println()

	// 百炼平台配置
	baiwanKey, ok := cfg.GetBaiwanAPIKey()
	if ok && baiwanKey != "" {
		fmt.Printf("✓ 百炼平台 API Key: 已配置 (%s)\n", config.MaskKey(baiwanKey))
	} else {
		fmt.Println("✗ 百炼平台 API Key: 未配置")
	}

	fmt.Println()

	// 服务可用性
	fmt.Println("服务可用性:")
	if tts.IsConfigured() {
		fmt.Println("  ✓ 语音合成 (TTS)")
	} else {
		fmt.Println("  ✗ 语音合成 (TTS) - 缺少百炼平台 API Key")
	}

	fmt.Println()

	// 运行模式
	mode, serverURL := cfg.GetMode()
	fmt.Println("运行模式:")
	if mode == "server" {
		fmt.Printf("  当前: 服务端中转 (%s)\n", serverURL)
	} else {
		fmt.Println("  当前: 直连百炼平台")
	}

	// 提示
	if !cfg.IsConfigured() {
		fmt.Println()
		fmt.Println("⚠️  配置不完整")
		fmt.Println("   请使用以下命令完成配置:")
		fmt.Println("     mini-tmk-agent config set-baiwan-key <api-key>")
	}

	return nil
}

func runClear(cmd *cobra.Command, args []string) error {
	cfg := config.GetInstance()

	if err := cfg.Clear(); err != nil {
		return fmt.Errorf("清除配置失败: %w", err)
	}

	fmt.Println("✅ 所有配置已清除")

	return nil
}

func runSetBaiwanKey(cmd *cobra.Command, args []string) error {
	apiKey := baiwanAPIKey
	if apiKey == "" && len(args) > 0 {
		apiKey = args[0]
	}

	if apiKey == "" {
		return fmt.Errorf("请提供百炼平台API Key")
	}

	if len(apiKey) < 10 {
		return fmt.Errorf("API Key格式不正确")
	}

	cfg := config.GetInstance()
	if err := cfg.SetBaiwanAPIKey(apiKey); err != nil {
		return fmt.Errorf("保存API Key失败: %w", err)
	}

	fmt.Println("✅ 百炼平台API Key已保存")
	fmt.Printf("   API Key: %s****\n", apiKey[:4])
	fmt.Println()
	fmt.Println("现在可以使用以下功能:")
	fmt.Println("  ✓ 语音识别 (ASR) - 实时语音转文字")
	fmt.Println("  ✓ 语音合成 (TTS) - 文字转语音")

	return nil
}

func runSetMode(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("请指定模式: server 或 direct\n用法: mini-tmk-agent config set-mode <server|direct>")
	}

	mode := args[0]
	if mode != "server" && mode != "direct" {
		return fmt.Errorf("无效的模式: %s\n有效模式: server (服务端中转) 或 direct (直连百炼平台)", mode)
	}

	cfg := config.GetInstance()

	// 直连模式需要检查百炼平台配置
	if mode == "direct" {
		if !cfg.IsConfigured() {
			fmt.Println()
			fmt.Println("⚠️  当前未配置百炼平台 API Key")
			fmt.Println()
			fmt.Println("请先配置百炼平台 API Key:")
			fmt.Println("  mini-tmk-agent config set-baiwan-key <api-key>")
			fmt.Println()
			return fmt.Errorf("直连模式需要配置百炼平台 API Key")
		}
	}

	serverURL := serverURLFlag
	if mode == "server" && serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	if err := cfg.SetMode(mode, serverURL); err != nil {
		return fmt.Errorf("保存模式配置失败: %w", err)
	}

	fmt.Println("✅ 模式已切换")
	if mode == "server" {
		fmt.Printf("   当前模式: 服务端中转 (%s)\n", serverURL)
		fmt.Println()
		fmt.Println("启动命令:")
		fmt.Println("  mini-tmk-agent stream --source-lang zh --target-lang en")
	} else {
		fmt.Println("   当前模式: 直连百炼平台")
		fmt.Println()
		fmt.Println("启动命令:")
		fmt.Println("  mini-tmk-agent stream --source-lang zh --target-lang en --direct")
	}

	return nil
}

func runGetMode(cmd *cobra.Command, args []string) error {
	cfg := config.GetInstance()
	mode, serverURL := cfg.GetMode()

	fmt.Println("运行模式")
	fmt.Println("========")
	fmt.Println()

	if mode == "server" {
		fmt.Println("当前模式: 服务端中转")
		fmt.Printf("服务端地址: %s\n", serverURL)
		fmt.Println()
		fmt.Println("说明: 请求通过服务端处理")
	} else {
		fmt.Println("当前模式: 直连百炼平台")
		fmt.Println()
		fmt.Println("说明: 本地直接调用百炼平台 API")

		// 显示配置状态
		if cfg.IsConfigured() {
			baiwanKey, _ := cfg.GetBaiwanAPIKey()
			fmt.Printf("百炼 API Key: 已配置 (%s)\n", config.MaskKey(baiwanKey))
		} else {
			fmt.Println("百炼 API Key: ⚠️ 未配置（请先运行配置）")
		}
	}

	fmt.Println()
	fmt.Println("切换模式:")
	fmt.Println("  mini-tmk-agent config set-mode server   # 切换到服务端中转模式")
	fmt.Println("  mini-tmk-agent config set-mode direct   # 切换到直连百炼平台模式")

	return nil
}
