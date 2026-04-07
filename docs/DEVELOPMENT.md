# MiniTMK Agent 开发文档

## 开发环境

### 前置要求

- Go 1.21+
- Git
- Make (可选)

### 克隆仓库

```bash
git clone https://github.com/luoleixi/MiniTMKAgent.git
cd MiniTMKAgent
```

### 安装依赖

```bash
go mod download
go mod tidy
```

---

## 项目结构

```
MiniTMKAgent/
├── cmd/                    # CLI 命令
│   ├── root.go            # 根命令
│   ├── interactive.go     # 交互式模式
│   ├── config.go          # 配置管理
│   └── update.go          # 更新功能
├── internal/              # 内部包
│   ├── agent/             # 业务逻辑
│   │   └── stream.go      # 流式同传 Agent
│   ├── audio/             # 音频处理
│   │   ├── player.go      # 播放器
│   │   ├── recorder.go    # 录音器
│   │   ├── vad.go         # 语音活动检测
│   │   └── play_queue.go  # 播放队列
│   ├── config/            # 配置管理
│   ├── recognizer/        # 语音识别
│   ├── translator/        # 翻译
│   ├── tts/               # 语音合成
│   └── utils/             # 工具函数
├── scripts/               # 安装脚本
│   ├── install.ps1        # Windows 安装
│   └── install.sh         # macOS/Linux 安装
├── docs/                  # 文档
├── main.go                # 入口
├── go.mod                 # Go 模块
└── README.md              # 项目说明
```

---

## 构建

### 本地构建

```bash
go build -o mini-tmk-agent .
```

### 交叉编译

**Windows:**
```bash
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o mini-tmk-agent.exe .
```

**macOS:**
```bash
# Intel
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o mini-tmk-agent-darwin-amd64 .

# Apple Silicon
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o mini-tmk-agent-darwin-arm64 .
```

**Linux:**
```bash
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o mini-tmk-agent-linux-amd64 .
```

---

## 核心模块说明

### 1. Agent 模块

`internal/agent/stream.go`

**职责:**
- 协调录音、识别、翻译、TTS 全流程
- 管理播放队列顺序
- 处理错误和超时

**关键类型:**
```go
type StreamAgent struct {
    config      StreamConfig
    recognizer  recognizer.Recognizer
    translator  translator.Translator
    tts         tts.TTS
    seqGen      *audio.SequenceGenerator
}
```

### 2. TTS 模块

`internal/tts/baiwan.go`

**职责:**
- 阿里云百炼平台 TTS 接口
- WebSocket 连接管理

**接口:**
```go
type TTS interface {
    Synthesize(text, voice string) ([]byte, error)
    SynthesizeStream(text, voice string, onAudioChunk func(chunk []byte)) error
    Close() error
}
```

### 3. 音频模块

`internal/audio/`

**播放队列:**
- 序列号生成器确保播放顺序
- 乱序音频缓存等待

**关键函数:**
```go
func GetPlayQueue() *PlayQueue
func (pq *PlayQueue) Enqueue(seq int, data []byte)
func (pq *PlayQueue) EnqueueWithoutSeq(data []byte) int
```

---

## 添加新功能

### 添加新命令

在 `cmd/` 目录创建新文件:

```go
package cmd

import "github.com/spf13/cobra"

var newCmd = &cobra.Command{
    Use:   "newcmd",
    Short: "新命令描述",
    RunE:  runNewCmd,
}

func init() {
    rootCmd.AddCommand(newCmd)
}

func runNewCmd(cmd *cobra.Command, args []string) error {
    // 实现逻辑
    return nil
}
```

### 添加新语言支持

1. **更新语言映射** (`internal/utils/lang.go`):
```go
var validLangCodes = map[string]string{
    "zh": "中文",
    "en": "英文",
    // 添加新语言
    "it": "意大利文",
}
```

2. **更新音色映射** (`internal/tts/interface.go`):
```go
var LanguageVoiceMap = map[string]string{
    "it": "longanyang", // 使用多语言通用音色
}
```

3. **添加快捷命令** (`cmd/interactive.go`):
```go
case "/start-it":
    startStreaming("zh", "it")
```

### 添加新的 TTS 提供商

1. 创建新文件 `internal/tts/new_provider.go`
2. 实现 `TTS` 接口
3. 在 `factory.go` 中注册

---

## 测试

### 运行测试

```bash
go test ./...
```

### 手动测试

```bash
# 构建并运行
go build -o mini-tmk-agent . && ./mini-tmk-agent
```

---

## 调试

### 启用调试日志

在代码中添加:
```go
import "mini-tmk-agent/internal/utils"

utils.Debug("调试信息")
utils.Debugf("格式化调试: %s", value)
```

### 日志位置

`logs/` 目录下按日期生成日志文件。

---

## 发布

### 版本号

使用语义化版本: `v1.0.0`

### 创建 Release

1. 更新版本号
2. 构建所有平台二进制文件
3. 创建 GitHub Release
4. 上传二进制文件

### 构建脚本

```bash
#!/bin/bash
VERSION="v1.0.0"

# Windows
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o mini-tmk-agent-${VERSION}-windows-amd64.exe .

# macOS
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o mini-tmk-agent-${VERSION}-darwin-amd64 .
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o mini-tmk-agent-${VERSION}-darwin-arm64 .

# Linux
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o mini-tmk-agent-${VERSION}-linux-amd64 .
```

---

## 代码规范

### 命名规范

- 包名: 小写，无下划线 (`audio`, `tts`)
- 接口名: 动词或形容词 (`TTS`, `Recognizer`)
- 结构体名: 名词 (`StreamAgent`, `PlayQueue`)
- 函数名: 动词开头 (`NewPlayer`, `Synthesize`)

### 错误处理

```go
if err != nil {
    return fmt.Errorf("操作失败: %w", err)
}
```

### 并发安全

```go
type SafeStruct struct {
    data  string
    mu    sync.Mutex
}

func (s *SafeStruct) Set(data string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.data = data
}
```

---

## 贡献指南

### 提交 Issue

- 描述问题
- 提供复现步骤
- 提供环境信息

### 提交 PR

1. Fork 仓库
2. 创建分支: `git checkout -b feature/xxx`
3. 提交更改: `git commit -m "Add xxx"`
4. 推送分支: `git push origin feature/xxx`
5. 创建 Pull Request

---

## 相关资源

- [Go 官方文档](https://golang.org/doc/)
- [Cobra 文档](https://github.com/spf13/cobra)
- [阿里云百炼平台](https://dashscope.console.aliyun.com/)
