package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"mini-tmk-agent/internal/agent"
	"mini-tmk-agent/internal/config"
	"mini-tmk-agent/internal/setup"
)

// 默认语言设置
var (
	defaultSourceLang = "zh"
	defaultTargetLang = "en"
	defaultDirectMode = false
)

var rootCmd = &cobra.Command{
	Use:   "mini-tmk-agent",
	Short: "简易同声传译Agent",
	Long: `mini-tmk-agent 是一个简易的同声传译Agent。

主要命令：
  - interactive: 交互式CLI模式（推荐，功能最全）
  - config: 配置管理
  - update: 检查并更新到最新版本

详细功能请使用 interactive 命令查看。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 双击运行（没有参数）时，进入 interactive 模式（显示菜单）
		if err := runInteractive(cmd, args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			fmt.Println("\n按任意键退出...")
			fmt.Scanln()
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runStreamDefault(cmd *cobra.Command, args []string) error {
	// 首次启动检查配置
	if err := setup.CheckAndSetup(); err != nil {
		return fmt.Errorf("配置失败: %w", err)
	}

	// 创建上下文，支持信号中断
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 捕获中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n🛑 接收到停止信号，正在关闭...")
		cancel()
	}()

	// 从配置读取模式
	cfg := config.GetInstance()
	mode, serverURL := cfg.GetMode()
	isDirectMode := mode == "direct"

	// 创建流式Agent（使用默认设置）
	streamAgent := agent.NewStreamAgent(agent.StreamConfig{
		SourceLang: defaultSourceLang,
		TargetLang: defaultTargetLang,
		DirectMode: isDirectMode,
		ServerURL:  serverURL,
	})

	// 启动流式翻译
	if err := streamAgent.Start(ctx); err != nil {
		return fmt.Errorf("流式翻译失败: %w", err)
	}

	return nil
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
