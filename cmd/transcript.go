package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"mini-tmk-agent/internal/agent"
)

var (
	audioFile   string
	outputFile  string
	transcriptLang string
)

var transcriptCmd = &cobra.Command{
	Use:   "transcript",
	Short: "转录模式",
	Long: `转录模式：提供一个本地音频文件（如 PCM、WAV 或 MP3 格式），
在本地生成转录好的文本文件。`,
	Example: "mini-tmk-agent transcript --file audio.wav --output transcript.txt --lang zh",
	RunE:    runTranscript,
}

func init() {
	rootCmd.AddCommand(transcriptCmd)
	transcriptCmd.Flags().StringVar(&audioFile, "file", "", "音频文件路径 (支持 PCM, WAV, MP3)")
	transcriptCmd.Flags().StringVar(&outputFile, "output", "", "输出文本文件路径")
	transcriptCmd.Flags().StringVar(&transcriptLang, "lang", "zh", "音频语言 (zh, en, ja, ko, fr, de, es, ru)")
	transcriptCmd.MarkFlagRequired("file")
	transcriptCmd.MarkFlagRequired("output")
}

func runTranscript(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// 创建转录Agent
	transcriptAgent := agent.NewTranscriptAgent(agent.TranscriptConfig{
		AudioFile:  audioFile,
		OutputFile: outputFile,
		Language:   transcriptLang,
	})

	// 执行转录
	if err := transcriptAgent.Run(ctx); err != nil {
		return fmt.Errorf("转录失败: %w", err)
	}

	return nil
}
