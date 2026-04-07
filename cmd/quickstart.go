package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"mini-tmk-agent/internal/config"
)

var quickstartCmd = &cobra.Command{
	Use:   "quickstart",
	Short: "一键快速启动",
	Long: `快速启动 MiniTMK Agent，自动完成配置。

支持以下方式配置 API Key：
  1. 命令行参数: mini-tmk-agent quickstart --api-key <your-key>
  2. 环境变量: export MINI_TMK_API_KEY=<your-key>
  3. 交互式输入: 直接运行命令后按提示输入

示例:
  # 使用命令行参数（推荐）
  mini-tmk-agent quickstart --api-key sk-xxxxxx

  # 使用环境变量
  export MINI_TMK_API_KEY=sk-xxxxxx
  mini-tmk-agent quickstart

  # 交互式配置（首次使用）
  mini-tmk-agent quickstart`,
	RunE: runQuickstart,
}

var quickstartAPIKey string

func init() {
	rootCmd.AddCommand(quickstartCmd)
	quickstartCmd.Flags().StringVarP(&quickstartAPIKey, "api-key", "k", "", "百炼平台 API Key")
}

func runQuickstart(cmd *cobra.Command, args []string) error {
	cfg := config.GetInstance()

	// 1. 获取 API Key（优先级：命令行参数 > 环境变量 > 已有配置 > 交互式输入）
	apiKey := quickstartAPIKey
	if apiKey == "" {
		apiKey = os.Getenv("MINI_TMK_API_KEY")
	}
	if apiKey == "" {
		if key, ok := cfg.GetBaiwanAPIKey(); ok {
			apiKey = key
		}
	}

	// 2. 如果没有 API Key，提示用户
	if apiKey == "" {
		fmt.Println()
		fmt.Println("╔══════════════════════════════════════════════════════════╗")
		fmt.Println("║           MiniTMK Agent - 快速启动向导                   ║")
		fmt.Println("╚══════════════════════════════════════════════════════════╝")
		fmt.Println()
		fmt.Println("首次使用需要配置百炼平台 API Key")
		fmt.Println()
		fmt.Println("获取方式:")
		fmt.Println("  1. 访问 https://dashscope.console.aliyun.com/")
		fmt.Println("  2. 进入 'API Key管理' 创建 API Key")
		fmt.Println()
		fmt.Print("请输入 API Key: ")
		fmt.Scanln(&apiKey)
		if apiKey == "" {
			return fmt.Errorf("API Key 不能为空，启动失败")
		}
		// 保存配置
		if err := cfg.SetBaiwanAPIKey(apiKey); err != nil {
			return fmt.Errorf("保存配置失败: %w", err)
		}
		if err := cfg.SetMode("direct", ""); err != nil {
			return fmt.Errorf("设置模式失败: %w", err)
		}
		fmt.Println("✅ API Key 已保存")
		fmt.Println()
	}

	// 3. 保存配置
	if err := cfg.SetBaiwanAPIKey(apiKey); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	// 4. 设置直连模式
	if err := cfg.SetMode("direct", ""); err != nil {
		return fmt.Errorf("设置模式失败: %w", err)
	}

	// 5. 显示启动信息
	fmt.Println()
	fmt.Println("✅ 配置完成！")
	fmt.Printf("   API Key: %s\n", config.MaskKey(apiKey))
	fmt.Println("   模式: 直连百炼平台")
	fmt.Println()
	fmt.Println("正在启动流式同传...")
	fmt.Println("   源语言: zh (中文)")
	fmt.Println("   目标语言: en (英文)")
	fmt.Println("   按 Ctrl+C 停止")
	fmt.Println()

	// 6. 启动流式同传
	return startStreamingWithDefaults()
}


// startStreamingWithDefaults 使用默认设置启动流式同传
func startStreamingWithDefaults() error {
	// 使用默认参数直接执行流式同传
	sourceLang = "zh"
	targetLang = "en"
	directMode = true
	return runStream(nil, nil)
}
