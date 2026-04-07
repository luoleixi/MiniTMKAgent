# MiniTMK Agent

简易同声传译 Agent，支持实时语音翻译和文件转录。

## 🚀 开箱即用（一行命令）

**Windows (PowerShell):**
```powershell
iwr -useb https://raw.githubusercontent.com/luoleixi/MiniTMKAgent/main/scripts/install.ps1 | iex
```

**macOS / Linux:**
```bash
curl -fsSL https://raw.githubusercontent.com/luoleixi/MiniTMKAgent/main/scripts/install.sh | bash
```

安装完成后，**重新打开终端**，即可使用：
```bash
mini-tmk-agent
```

---

## 📦 其他安装方式

### 预编译二进制

从 [Releases 页面](https://github.com/luoleixi/MiniTMKAgent/releases) 下载对应系统的版本，解压后使用：

```bash
# 直接运行进入交互式模式
./mini-tmk-agent
```

### 从源码构建

```bash
git clone https://github.com/luoleixi/MiniTMKAgent.git
cd MiniTMKAgent
go build -o mini-tmk-agent .
./mini-tmk-agent
```

---

## 🔧 功能说明（交互式模式）

运行 `mini-tmk-agent` 进入交互式 CLI，支持以下命令：

### 实时同传

```
> /start              # 启动同传（默认 zh → en）
> /start zh en        # 中文 → 英文
> /start en zh        # 英文 → 中文
> /start zh ja        # 中文 → 日文
> /start-en           # 快捷：中文 → 英文
> /start-ja           # 快捷：中文 → 日文
> /start-ko           # 快捷：中文 → 韩文
```

**支持的语言：** zh（中文）、en（英文）、ja（日文）、ko（韩文）、fr（法文）、de（德文）、es（西班牙文）、ru（俄文）

### 文件转录

```
> /transcript audio.wav output.txt zh    # 转录音频为文本
```

支持格式：wav, mp3, pcm, m4a, flac, aac, ogg

### 其他命令

```
> /config        # 查看/修改配置
> /help          # 显示帮助
> /quit          # 退出程序
```

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

## 📝 顶层命令参考

```bash
# 进入交互式模式（推荐）
mini-tmk-agent

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
