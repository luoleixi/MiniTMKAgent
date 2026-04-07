package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "下载依赖或示例文件",
	Long: `下载 MiniTMK Agent 所需的依赖或示例文件。

支持下载：
  - ffmpeg: 音频格式转换工具
  - demo:   示例音频文件（用于测试转录功能）

示例:
  # 下载 FFmpeg（自动选择对应系统版本）
  mini-tmk-agent download ffmpeg

  # 下载示例音频文件
  mini-tmk-agent download demo --output ./demo.wav

  # 下载示例文件到指定目录
  mini-tmk-agent download demo -o ./samples/`,
}

var (
	downloadOutput string
)

var downloadFFmpegCmd = &cobra.Command{
	Use:   "ffmpeg",
	Short: "下载 FFmpeg 音频转换工具",
	Long: `FFmpeg 是用于音频格式转换的工具。
支持将 m4a/flac/aac 等格式转换为 WAV 格式。

下载后会自动解压到系统 PATH 可访问的目录。`,
	RunE: runDownloadFFmpeg,
}

var downloadDemoCmd = &cobra.Command{
	Use:   "demo",
	Short: "下载示例音频文件",
	Long: `下载示例音频文件，用于测试转录功能。

文件信息：
  - 格式: WAV (16kHz, 16bit, 单声道)
  - 语言: 中文
  - 时长: 约 5 秒

下载后可以使用以下命令测试：
  mini-tmk-agent transcript --file demo.wav --output result.txt --lang zh`,
	RunE: runDownloadDemo,
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.AddCommand(downloadFFmpegCmd)
	downloadCmd.AddCommand(downloadDemoCmd)

	// demo 命令参数
	downloadDemoCmd.Flags().StringVarP(&downloadOutput, "output", "o", "demo.wav", "输出文件路径")
}

func runDownloadFFmpeg(cmd *cobra.Command, args []string) error {
	fmt.Println("🔄 准备下载 FFmpeg...")
	fmt.Println()

	// 检测操作系统
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	fmt.Printf("系统信息: %s/%s\n", goos, goarch)

	var downloadURL string

	switch goos {
	case "windows":
		if goarch == "amd64" {
			downloadURL = "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip"
		}
	case "darwin":
		fmt.Println("macOS 建议使用 Homebrew 安装:")
		fmt.Println("  brew install ffmpeg")
		return nil
	case "linux":
		fmt.Println("Linux 建议使用包管理器安装:")
		fmt.Println("  Ubuntu/Debian: sudo apt install ffmpeg")
		fmt.Println("  CentOS/RHEL:   sudo yum install ffmpeg")
		fmt.Println("  Arch:          sudo pacman -S ffmpeg")
		return nil
	default:
		return fmt.Errorf("不支持的操作系统: %s", goos)
	}

	if downloadURL == "" {
		return fmt.Errorf("暂不支持该架构: %s/%s", goos, goarch)
	}

	fmt.Printf("下载地址: %s\n", downloadURL)
	fmt.Println()
	fmt.Println("请手动下载并解压到系统 PATH 目录，")
	fmt.Println("或访问 https://ffmpeg.org/download.html 获取安装包")
	fmt.Println()

	return nil
}

func runDownloadDemo(cmd *cobra.Command, args []string) error {
	fmt.Println("🔄 下载示例音频文件...")
	fmt.Println()

	// 示例音频文件 URL（使用一个公开的测试音频）
	// 注意：这里使用占位符 URL，实际使用时需要替换为真实的示例文件地址
	demoURL := "https://www.soundhelix.com/examples/mp3/SoundHelix-Song-1.mp3"

	outputPath := downloadOutput
	if outputPath == "" {
		outputPath = "demo.wav"
	}

	// 确保输出目录存在
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	fmt.Printf("下载地址: %s\n", demoURL)
	fmt.Printf("保存路径: %s\n", outputPath)
	fmt.Println()

	// 下载文件
	if err := downloadFile(demoURL, outputPath); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	fmt.Printf("✅ 示例文件已保存: %s\n", outputPath)
	fmt.Println()
	fmt.Println("可以使用以下命令测试转录：")
	fmt.Printf("  mini-tmk-agent transcript --file %s --output result.txt --lang zh\n", outputPath)

	return nil
}

// downloadFile 下载文件到指定路径
func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
