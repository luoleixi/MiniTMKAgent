# MiniTMK Agent

简易同声传译 Agent，支持实时语音翻译和文件转录。

## 🚀 开箱即用（一行命令）

**Windows PowerShell:**
```powershell
iwr -useb https://raw.githubusercontent.com/luoleixi/MiniTMKAgent/main/scripts/install.ps1 | iex
```

安装完成后，**重新打开终端**，即可使用：
```powershell
mini-tmk-agent quickstart
```

---

## 📦 其他安装方式

### 预编译二进制

从 [Releases 页面](../../releases) 下载对应系统的版本，解压后使用：

```powershell
# Windows
.\mini-tmk-agent.exe quickstart

# macOS/Linux
./mini-tmk-agent quickstart
```

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/luoleixi/MiniTMKAgent.git
cd MiniTMKAgent

# 构建
go build -o mini-tmk-agent.exe .

# 运行
.\mini-tmk-agent.exe quickstart
```

---

## 🔧 功能说明

### 实时同传模式

```bash
# 中文 → 英文
mini-tmk-agent stream --source-lang zh --target-lang en

# 英文 → 中文
mini-tmk-agent stream --source-lang en --target-lang zh

# 日文 → 中文
mini-tmk-agent stream --source-lang ja --target-lang zh
```

**支持的语言：** zh（中文）、en（英文）、ja（日文）、ko（韩文）、fr（法文）、de（德文）、es（西班牙文）、ru（俄文）

### 文件转录模式

```bash
# 转录音频文件为文本
mini-tmk-agent transcript --file audio.wav --output result.txt --lang zh

# 支持格式：wav, mp3, pcm, m4a, flac, aac, ogg
```

### 交互式 CLI 模式

```bash
mini-tmk-agent interactive
```

常用命令：
- `/start` - 启动同传（默认 zh → en）
- `/start zh en` - 指定语言对
- `/transcript file.wav out.txt zh` - 转录音频
- `/help` - 显示帮助
- `/quit` - 退出

---

## ⚙️ 配置说明

### 获取百炼平台 API Key

1. 访问 [阿里云百炼平台](https://dashscope.console.aliyun.com/)
2. 进入 "API Key 管理"
3. 创建 API Key
4. 复制并使用

### 配置文件位置

- **Windows:** `%APPDATA%\mini-tmk-agent\config.json`
- **macOS:** `~/Library/Application Support/mini-tmk-agent/config.json`
- **Linux:** `~/.config/mini-tmk-agent/config.json`

---

## 📝 命令参考

```bash
# 查看帮助
mini-tmk-agent --help

# 配置管理
mini-tmk-agent config status              # 查看配置状态
mini-tmk-agent config set-baiwan-key      # 设置 API Key
mini-tmk-agent config set-mode direct     # 设置直连模式

# 自动更新
mini-tmk-agent update                     # 检查并更新到最新版本
mini-tmk-agent update --check             # 仅检查更新
```

---

## 📄 许可证

MIT License

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！
