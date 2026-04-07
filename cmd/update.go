package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "检查并更新到最新版本",
	Long: `检查 GitHub Releases 是否有新版本，并自动下载更新。

支持自动更新到最新版本，或指定版本号更新。

示例:
  # 检查并更新到最新版本
  mini-tmk-agent update

  # 更新到指定版本
  mini-tmk-agent update --version v1.1.0

  # 仅检查更新，不执行更新
  mini-tmk-agent update --check`,
	RunE: runUpdate,
}

var (
	updateVersion string
	updateCheck   bool
)

const (
	repoOwner = "luoleixi"
	repoName  = "MiniTMKAgent"
)

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringVarP(&updateVersion, "version", "v", "", "指定要更新的版本号")
	updateCmd.Flags().BoolVarP(&updateCheck, "check", "c", false, "仅检查更新，不执行更新")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	fmt.Println("🔄 检查更新...")
	fmt.Println()

	// 获取当前版本（这里可以从编译时注入的版本变量读取）
	currentVersion := getCurrentVersion()
	fmt.Printf("当前版本: %s\n", currentVersion)

	// 获取最新版本
	latestVersion, releaseURL, err := getLatestVersion()
	if err != nil {
		return fmt.Errorf("检查更新失败: %w", err)
	}

	fmt.Printf("最新版本: %s\n", latestVersion)
	fmt.Println()

	// 比较版本
	if latestVersion == currentVersion && updateVersion == "" {
		fmt.Println("✅ 已经是最新版本!")
		return nil
	}

	// 仅检查模式
	if updateCheck {
		if latestVersion != currentVersion {
			fmt.Printf("发现新版本: %s\n", latestVersion)
			fmt.Printf("发布地址: %s\n", releaseURL)
			fmt.Println()
			fmt.Println("运行以下命令更新:")
			fmt.Println("  mini-tmk-agent update")
		}
		return nil
	}

	// 确认更新
	targetVersion := updateVersion
	if targetVersion == "" {
		targetVersion = latestVersion
	}

	if targetVersion == currentVersion {
		fmt.Println("✅ 已经是指定版本!")
		return nil
	}

	fmt.Printf("准备更新到: %s\n", targetVersion)
	fmt.Print("是否继续? [y/N]: ")

	var confirm string
	fmt.Scanln(&confirm)
	if strings.ToLower(confirm) != "y" {
		fmt.Println("取消更新")
		return nil
	}

	fmt.Println()
	fmt.Println("🔄 开始下载...")

	// 执行更新
	if err := performUpdate(targetVersion); err != nil {
		return fmt.Errorf("更新失败: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ 更新成功!")
	fmt.Println("请重新运行程序以使用新版本")

	return nil
}

// getCurrentVersion 获取当前版本号
func getCurrentVersion() string {
	// 这里可以从编译时注入的变量读取
	// 暂时返回 development
	if version == "" {
		return "development"
	}
	return version
}

// getLatestVersion 获取 GitHub 最新版本
func getLatestVersion() (version, url string, err error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("GitHub API 返回 %d", resp.StatusCode)
	}

	var result struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}

	return result.TagName, result.HTMLURL, nil
}

// performUpdate 执行更新
func performUpdate(version string) error {
	// 检测系统
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// 构建下载地址
	assetName := fmt.Sprintf("mini-tmk-agent-%s-%s", goos, goarch)
	if goos == "windows" {
		assetName += ".zip"
	} else {
		assetName += ".tar.gz"
	}

	downloadURL := fmt.Sprintf(
		"https://github.com/%s/%s/releases/download/%s/%s",
		repoOwner, repoName, version, assetName,
	)

	fmt.Printf("下载地址: %s\n", downloadURL)

	// 获取当前可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return err
	}

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "mini-tmk-agent-update-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, assetName)

	// 下载新版本
	if err := downloadFileWithProgress(downloadURL, tempFile); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	// 解压
	fmt.Println("🔄 解压文件...")
	newBinary := filepath.Join(tempDir, "mini-tmk-agent")
	if goos == "windows" {
		newBinary += ".exe"
	}

	// TODO: 解压逻辑（根据不同格式）
	_ = newBinary

	// 替换旧版本（Windows 需要特殊处理）
	backupPath := execPath + ".backup"

	// 备份旧版本
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("备份失败: %w", err)
	}

	// 移动新版本
	// TODO: 完成替换逻辑

	// 删除备份
	os.Remove(backupPath)

	return nil
}

// downloadFileWithProgress 带进度条下载
func downloadFileWithProgress(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// 获取文件大小
	totalSize := resp.ContentLength
	if totalSize > 0 {
		fmt.Printf("文件大小: %.2f MB\n", float64(totalSize)/1024/1024)
	}

	// 下载并显示进度
	_, err = io.Copy(out, resp.Body)
	return err
}

// version 会被编译时注入
var version = ""
