package audio

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// SystemPlayer 系统命令行音频播放器（用于播放MP3文件）
type SystemPlayer struct {
	playCmd string
	args    []string
}

// NewSystemPlayer 创建系统音频播放器
func NewSystemPlayer() (*SystemPlayer, error) {
	cmd, args := getSystemPlayerCommand()
	if cmd == "" {
		return nil, fmt.Errorf(getInstallHint())
	}

	return &SystemPlayer{
		playCmd: cmd,
		args:    args,
	}, nil
}

// PlayFile 播放音频文件
func (p *SystemPlayer) PlayFile(filePath string) error {
	// Windows 特殊处理 - 使用 Edge 浏览器播放（所有 Win10/11 都预装）
	if p.playCmd == "cmd" && len(p.args) >= 3 && p.args[2] == "msedge" {
		// 使用 cmd /c start msedge "filepath"
		args := []string{"/c", "start", "msedge", filePath}
		cmd := exec.Command(p.playCmd, args...)
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("播放失败: %w", err)
		}
		// 不等待，直接返回
		return nil
	}

	args := append(p.args, filePath)
	cmd := exec.Command(p.playCmd, args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("播放失败: %w", err)
	}
	return nil
}

// PlayMP3Data 播放MP3数据（保存到临时文件后播放）
func (p *SystemPlayer) PlayMP3Data(data []byte) error {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "tts_*.mp3")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// 写入音频数据
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("写入音频数据失败: %w", err)
	}
	tmpFile.Close()

	// 播放
	return p.PlayFile(tmpFile.Name())
}

// getSystemPlayerCommand 获取系统播放器命令
func getSystemPlayerCommand() (cmd string, args []string) {
	switch runtime.GOOS {
	case "darwin":
		// macOS 使用 afplay
		if _, err := exec.LookPath("afplay"); err == nil {
			return "afplay", nil
		}
	case "windows":
		// Windows 使用 Edge 浏览器播放（所有 Win10/11 都预装）
		// Edge 支持 MP3 播放，且不会阻塞
		return "cmd", []string{"/c", "start", "msedge"}
	case "linux":
		// Linux 尝试使用常见的播放器（按推荐程度排序）
		if _, err := exec.LookPath("mpv"); err == nil {
			return "mpv", []string{"--no-video", "--really-quiet"}
		}
		if _, err := exec.LookPath("cvlc"); err == nil {
			return "cvlc", []string{"--play-and-exit", "--quiet"}
		}
		if _, err := exec.LookPath("ffplay"); err == nil {
			return "ffplay", []string{"-nodisp", "-autoexit", "-loglevel", "quiet"}
		}
		if _, err := exec.LookPath("mpg123"); err == nil {
			return "mpg123", []string{"-q"}
		}
		if _, err := exec.LookPath("paplay"); err == nil {
			return "paplay", nil
		}
	}
	return "", nil
}

// IsSystemPlayerAvailable 检查是否有可用的系统播放器
func IsSystemPlayerAvailable() bool {
	cmd, _ := getSystemPlayerCommand()
	return cmd != ""
}

// getInstallHint 获取安装提示
func getInstallHint() string {
	switch runtime.GOOS {
	case "darwin":
		return "macOS 系统缺少音频播放器，请安装:\n" +
			"  brew install mpv    # 推荐\n" +
			"  或\n" +
			"  brew install ffmpeg # 包含 ffplay"
	case "windows":
		return "Windows 音频播放需要系统组件，请确保系统更新到最新版本"
	case "linux":
		return "Linux 系统缺少音频播放器，请安装:\n" +
			"  Ubuntu/Debian: sudo apt install mpv vlc\n" +
			"  CentOS/RHEL:   sudo yum install mpv vlc\n" +
			"  Arch Linux:    sudo pacman -S mpv vlc"
	default:
		return "未找到可用的音频播放器，请安装 mpv、ffplay 或 VLC"
	}
}

// PlayMP3File 播放MP3文件（便捷函数）
// 优先使用纯 Go 播放器，无需外部依赖
func PlayMP3File(filePath string) error {
	// 读取文件数据
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}
	// 使用纯 Go 播放器播放
	return PlayMP3Data(data)
}

// PlayMP3Data 播放MP3数据（便捷函数）
// 使用播放队列确保多条语音按顺序播放
func PlayMP3Data(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	// 使用播放队列，确保按顺序播放
	queue := GetPlayQueue()
	queue.Enqueue(data)
	return nil
}

// PlayMP3DataDirect 直接播放MP3数据（不使用队列，会打断当前播放）
// 优先使用纯 Go 播放器，无需外部依赖
func PlayMP3DataDirect(data []byte) error {
	// 尝试使用纯 Go 播放器（推荐，无外部依赖）
	goPlayer, err := GetGoPlayer()
	if err == nil {
		return goPlayer.PlayMP3Data(data)
	}

	// 回退到系统播放器（备用方案）
	player, err := NewSystemPlayer()
	if err != nil {
		return err
	}
	return player.PlayMP3Data(data)
}
