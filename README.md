# MiniTMK Agent

简易同声传译 Agent，支持实时语音翻译和文件转录。

## 🚀 快速开始（开箱即用）

### 方式一：一行命令安装（推荐）

**macOS / Linux:**
```bash
curl -fsSL https://raw.githubusercontent.com/luoleixi/mini-tmk-agent/main/scripts/install.sh | bash
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/luoleixi/mini-tmk-agent/main/scripts/install.ps1 | iex
```

或者使用完整命令：
```powershell
Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/luoleixi/mini-tmk-agent/main/scripts/install.ps1' -UseBasicParsing | Invoke-Expression
```

安装完成后直接运行：
```bash
mini-tmk-agent quickstart
```

---

### 方式二：使用预编译二进制文件

1. **下载对应系统的版本** [Releases 页面](../../releases)

2. **解压后双击运行或直接命令行启动**

   **Windows:**
   ```powershell
   # 快速启动（交互式输入 API Key）
   .\mini-tmk-agent.exe quickstart
   
   # 或直接带 API Key 启动
   .\mini-tmk-agent.exe quickstart --api-key sk-xxxxxx
   ```

   **macOS/Linux:**
   ```bash
   # 快速启动
   ./mini-tmk-agent quickstart
   
   # 或带 API Key 启动
   ./mini-tmk-agent quickstart --api-key sk-xxxxxx
   ```

3. **开始使用**
   - 程序自动配置并启动
   - 对着麦克风说话，自动实时翻译
   - 按 `Ctrl+C` 停止

### 方式三：使用环境变量

```bash
# 设置 API Key（只需一次）
export MINI_TMK_API_KEY=sk-xxxxxx  # Linux/macOS
set MINI_TMK_API_KEY=sk-xxxxxx     # Windows CMD
$env:MINI_TMK_API_KEY="sk-xxxxxx"  # PowerShell

# 启动
mini-tmk-agent quickstart
```

### 方式四：交互式配置

```bash
mini-tmk-agent quickstart
# 按提示输入百炼平台 API Key
```

---

## 📦 安装

### 一行命令安装

**macOS / Linux:**
```bash
curl -fsSL https://raw.githubusercontent.com/luoleixi/MiniTMKAgent/main/scripts/install.sh | bash
```

**Windows (PowerShell 管理员):**
```powershell
powershell -ExecutionPolicy Bypass -Command "Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/luoleixi/MiniTMKAgent/main/scripts/install.ps1' -UseBasicParsing | Invoke-Expression"
```

安装脚本会自动：
- 检测系统架构
- 下载最新版本
- 安装到系统 PATH
- 创建快速启动脚本

### 从 Release 下载

| 系统 | 架构 | 下载 |
|------|------|------|
| Windows | amd64 | [下载](../../releases/latest) |
| macOS | amd64/arm64 | [下载](../../releases/latest) |
| Linux | amd64/arm64 | [下载](../../releases/latest) |

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/luoleixi/mini-tmk-agent.git
cd mini-tmk-agent

# 构建
go build -o mini-tmk-agent .

# 运行
./mini-tmk-agent quickstart
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

# 下载依赖/示例
mini-tmk-agent download ffmpeg            # FFmpeg 安装指南
mini-tmk-agent download demo              # 下载示例音频

# 自动更新
mini-tmk-agent update                     # 检查并更新到最新版本
mini-tmk-agent update --check             # 仅检查更新
mini-tmk-agent update --version v1.1.0    # 更新到指定版本
```

---

## 📄 许可证

MIT License

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！