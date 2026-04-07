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

var (
	sourceLang string
	targetLang string
	directMode bool
)

var streamCmd = &cobra.Command{
	Use:   "stream",
	Short: "流式同传模式",
	Long: `流式同传模式：持续监听电脑麦克风，一旦检测到有源目标语言输入，
会实时在终端上同时显示源目标语言识别后的文字和翻译后目标语言的文字。

支持的语言代码:
  zh = 中文      en = English   ja = 日本語
  ko = 한국어    fr = Français  de = Deutsch
  es = Español   ru = Русский`,
	Example: `  mini-tmk-agent stream --source-lang zh --target-lang en
  mini-tmk-agent stream --source-lang en --target-lang zh
  mini-tmk-agent stream --source-lang ja --target-lang zh`,
	RunE: runStream,
}

func init() {
	// stream 命令已移除，请使用 interactive 命令
	// rootCmd.AddCommand(streamCmd)
	// streamCmd.Flags().StringVar(&sourceLang, "source-lang", "zh", "源语言代码 (zh, en, ja, ko, fr, de, es, ru)")
	// streamCmd.Flags().StringVar(&targetLang, "target-lang", "en", "目标语言代码 (zh, en, ja, ko, fr, de, es, ru)")
	// streamCmd.Flags().BoolVar(&directMode, "direct", false, "使用直连模式（绕过服务端中转）")
}

func runStream(cmd *cobra.Command, args []string) error {
	// 首次启动检查配置（在信号处理之前，避免配置过程中 Ctrl+C 导致混乱）
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

	// 如果命令行指定了 --direct，覆盖配置
	if directMode {
		isDirectMode = true
	}

	// 创建流式Agent
	streamAgent := agent.NewStreamAgent(agent.StreamConfig{
		SourceLang: sourceLang,
		TargetLang: targetLang,
		DirectMode: isDirectMode,
		ServerURL:  serverURL,
	})

	// 启动流式翻译
	if err := streamAgent.Start(ctx); err != nil {
		return fmt.Errorf("流式翻译失败: %w", err)
	}

	return nil
}
