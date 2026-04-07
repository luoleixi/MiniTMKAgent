# MiniTMK Agent

<p align="center">
  <strong>简易同声传译 Agent</strong><br>
  支持实时语音翻译和文件转录
</p>

<p align="center">
  <a href="https://github.com/luoleixi/MiniTMKAgent/releases">
    <img src="https://img.shields.io/github/v/release/luoleixi/MiniTMKAgent" alt="Release">
  </a>
  <a href="LICENSE">
    <img src="https://img.shields.io/github/license/luoleixi/MiniTMKAgent" alt="License">
  </a>
</p>

## 快速开始

### 一键安装

**Windows (PowerShell):**
```powershell
iwr -useb https://raw.githubusercontent.com/luoleixi/MiniTMKAgent/main/scripts/install.ps1 | iex
```

**macOS / Linux:**
```bash
curl -fsSL https://raw.githubusercontent.com/luoleixi/MiniTMKAgent/main/scripts/install.sh | bash
```

安装完成后，**重新打开终端**：
```bash
mini-tmk-agent
```

### 首次使用

1. 运行 `mini-tmk-agent`
2. 按提示输入 [阿里云百炼平台](https://dashscope.console.aliyun.com/) API Key
3. 进入交互式菜单，输入 `/start` 开始同传

---

## 使用指南

### 交互式命令

```
> /start              # 启动同传（默认 zh → en）
> /start zh en        # 指定语言对
> /start-en           # 快捷：中文 → 英文
> /start-ja           # 快捷：中文 → 日文
> /transcript file.mp3 out.txt   # 转录音频文件
> /config             # 配置管理
> /help               # 显示帮助
> /quit               # 退出
```

### 顶层命令

```bash
mini-tmk-agent              # 进入交互式模式
mini-tmk-agent --help       # 查看帮助
mini-tmk-agent config       # 配置管理
mini-tmk-agent update       # 检查更新
```

详细文档：[用户指南](docs/USER_GUIDE.md)

---

## 文档

| 文档 | 描述 |
|------|------|
| [用户指南](docs/USER_GUIDE.md) | 安装、配置、使用说明 |
| [架构设计](docs/ARCHITECTURE.md) | 系统架构、模块设计 |
| [开发文档](docs/DEVELOPMENT.md) | 开发环境、构建、贡献指南 |

---

## 安装方式

### 方式一：脚本安装 (推荐)

见 [快速开始](#快速开始)

### 方式二：预编译二进制

从 [Releases](https://github.com/luoleixi/MiniTMKAgent/releases) 下载对应版本：

```bash
# Windows
.\mini-tmk-agent.exe

# macOS / Linux
./mini-tmk-agent
```

### 方式三：源码构建

```bash
git clone https://github.com/luoleixi/MiniTMKAgent.git
cd MiniTMKAgent
go build -o mini-tmk-agent .
./mini-tmk-agent
```

[//]: # (---)

[//]: # ()
[//]: # (## 系统要求)

[//]: # ()
[//]: # (| 平台 | 最低版本 | 架构 |)

[//]: # (|------|----------|------|)

[//]: # (| Windows | 10 | amd64 |)

[//]: # (| macOS | 10.15 | amd64, arm64 |)

[//]: # (| Linux | Ubuntu 18.04+ | amd64, arm64 |)

[//]: # ()
[//]: # (**依赖：**)

[//]: # (- 网络连接（阿里云百炼平台 API）)

[//]: # (- 麦克风权限)

[//]: # (- 音频输出设备)

---

[//]: # (## 项目架构)

[//]: # ()
[//]: # (```)

[//]: # (┌─────────────┐     ┌─────────────┐     ┌─────────────┐)

[//]: # (│   麦克风    │────▶│    VAD      │────▶│    ASR      │)

[//]: # (└─────────────┘     └─────────────┘     └──────┬──────┘)

[//]: # (                                                │)

[//]: # (                       ┌────────────────────────┘)

[//]: # (                       ▼)

[//]: # (┌─────────────┐     ┌─────────────┐     ┌─────────────┐)

[//]: # (│   扬声器    │◀────│  播放队列   │◀────│    TTS      │)

[//]: # (└─────────────┘     └─────────────┘     └─────────────┘)

[//]: # (```)

[//]: # ()
[//]: # (详细架构：[架构设计文档]&#40;docs/ARCHITECTURE.md&#41;)

[//]: # ()
[//]: # (---)

[//]: # ()
[//]: # (## 技术栈)

[//]: # ()
[//]: # (- **语言**: Go 1.21+)

[//]: # (- **音频**: miniaudio &#40;malgo&#41;, oto)

[//]: # (- **网络**: WebSocket, HTTP/2)

[//]: # (- **API**: 阿里云百炼平台)

[//]: # ()
[//]: # (---)

[//]: # (## 贡献)

[//]: # ()
[//]: # (欢迎提交 Issue 和 Pull Request！)

[//]: # ()
[//]: # (1. Fork 本仓库)

[//]: # (2. 创建特性分支 &#40;`git checkout -b feature/AmazingFeature`&#41;)

[//]: # (3. 提交更改 &#40;`git commit -m 'Add some AmazingFeature'`&#41;)

[//]: # (4. 推送分支 &#40;`git push origin feature/AmazingFeature`&#41;)

[//]: # (5. 创建 Pull Request)

[//]: # ()
[//]: # (详细指南：[开发文档]&#40;docs/DEVELOPMENT.md&#41;)

[//]: # ()
[//]: # (---)

## 许可证

MIT License - 详见 [LICENSE](LICENSE)

[//]: # (---)

[//]: # ()
[//]: # (## 致谢)

[//]: # ()
[//]: # (- [阿里云百炼平台]&#40;https://dashscope.console.aliyun.com/&#41; 提供语音识别、翻译、语音合成服务)

[//]: # (- [miniaudio]&#40;https://github.com/mackron/miniaudio&#41; 提供跨平台音频支持)

[//]: # ()
[//]: # (---)

<p align="center">
  Made with ❤️ by <a href="https://github.com/luoleixi">luoleixi</a>
</p>
