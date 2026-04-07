# MiniTMK Agent 用户指南

## 目录

1. [快速开始](#快速开始)
2. [安装](#安装)
3. [基本使用](#基本使用)
4. [交互式命令](#交互式命令)
5. [配置管理](#配置管理)
6. [常见问题](#常见问题)
7. [故障排除](#故障排除)

---

## 快速开始

### 一键安装 (推荐)

**Windows:**
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

---

## 安装

### 方式一：脚本安装 (推荐)

见 [快速开始](#快速开始)

### 方式二：手动下载

1. 访问 [Releases](https://github.com/luoleixi/MiniTMKAgent/releases) 页面
2. 下载对应系统的版本
3. 解压到任意目录
4. 添加到 PATH 或直接运行

```bash
# Windows
.\mini-tmk-agent.exe

# macOS / Linux
./mini-tmk-agent
```

### 方式三：源码构建

需要 Go 1.21+ 环境：

```bash
git clone https://github.com/luoleixi/MiniTMKAgent.git
cd MiniTMKAgent
go build -o mini-tmk-agent .
```

---

## 基本使用

### 启动程序

```bash
mini-tmk-agent
```

首次使用会提示输入 API Key。

### 首次配置

1. 访问 [阿里云百炼平台](https://dashscope.console.aliyun.com/)
2. 创建 API Key
3. 复制 Key 粘贴到程序中
4. 配置自动保存

---

## 交互式命令

### 实时同传

#### 启动同传
```
> /start
```
默认启动 **中文 → 英文** 同传。

#### 指定语言对
```
> /start zh en    # 中文 → 英文
> /start en zh    # 英文 → 中文
> /start zh ja    # 中文 → 日文
> /start ja zh    # 日文 → 中文
```

#### 快捷命令
```
> /start-en       # 快捷：中文 → 英文
> /start-ja       # 快捷：中文 → 日文
> /start-ko       # 快捷：中文 → 韩文
> /start-fr       # 快捷：中文 → 法文
> /start-de       # 快捷：中文 → 德文
> /start-es       # 快捷：中文 → 西班牙文
> /start-ru       # 快捷：中文 → 俄文
```

#### 停止同传
按 `Ctrl+C` 停止同传，返回交互式菜单。

### 文件转录

```
> /transcript audio.mp3 output.txt zh
```

参数：
- `audio.wav`: 音频文件路径
- `output.txt`: 输出文本文件
- `zh`: 音频语言 (可选，默认中文)

支持格式：MP3

### 配置管理

```
> /config
```

进入配置菜单：
1. 设置/修改 API Key
2. 切换直连/中继模式
3. 查看当前配置

### 帮助

```
> /help
```

显示所有可用命令。

### 退出

```
> /quit
```

退出程序。

---

## 配置管理

### 查看配置

```bash
mini-tmk-agent config status
```

### 设置 API Key

```bash
mini-tmk-agent config set-baiwan-key
```

### 配置文件位置

- **Windows:** `%APPDATA%\mini-tmk-agent\config.json`
- **macOS:** `~/Library/Application Support/mini-tmk-agent/config.json`
- **Linux:** `~/.config/mini-tmk-agent/config.json`

### 手动编辑配置

```json
{
  "mode": "direct",
  "server_url": "",
  "baiwan": {
    "api_key": "sk-xxxxxx"
  }
}
```

---

## 常见问题

### Q: 支持哪些语言？

**支持的语言：**
- 中文 (zh)
- 英文 (en)
- 日文 (ja)
- 韩文 (ko)
- 法文 (fr)
- 德文 (de)
- 西班牙文 (es)
- 俄文 (ru)

### Q: 需要联网吗？

是的，需要联网。所有语音和翻译服务都通过阿里云百炼平台 API 实现。

### Q: 数据安全吗？

- API Key 仅存储在本地配置文件
- 语音数据通过 HTTPS/WebSocket 加密传输
- 不收集任何用户数据

### Q: 支持离线使用吗？

不支持。需要连接阿里云百炼平台服务。

### Q: 可以自定义音色吗？

目前使用默认音色。如需自定义，请修改源码中的音色配置。

### Q: 延迟多少？

端到端延迟约 1-3 秒，取决于网络状况。

---

## 故障排除

### 问题：无法识别语音

**检查：**
1. 麦克风是否正常工作
2. 系统麦克风权限是否授予
3. 麦克风是否被其他程序占用

**解决：**
```bash
# Windows 检查麦克风
Settings > Privacy > Microphone > Allow apps to access your microphone

# macOS 检查麦克风
System Preferences > Security > Privacy > Microphone
```

### 问题：TTS 合成失败

**常见原因：**
- API Key 无效或过期
- 网络连接问题
- 文本包含不支持的特殊字符

**解决：**
1. 检查 API Key: `mini-tmk-agent config status`
2. 重新设置 API Key: `mini-tmk-agent config set-baiwan-key`
3. 检查网络连接

### 问题：播放没有声音

**检查：**
1. 扬声器音量是否开启
2. 默认音频输出设备是否正确
3. 程序是否有音频播放权限

**解决：**
- Windows: 检查音量混合器
- macOS: 检查系统偏好设置 > 声音
- Linux: 检查 `pavucontrol` 或 `alsamixer`

### 问题：程序崩溃

**收集信息：**
1. 运行程序时的完整错误信息
2. 操作系统版本
3. 程序版本 (`mini-tmk-agent --version`)

**报告问题：**
在 GitHub Issues 提交问题，附上以上信息。

### 问题：安装失败

**Windows:**
```powershell
# 检查 PowerShell 执行策略
Get-ExecutionPolicy
# 如需更改
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

**macOS/Linux:**
```bash
# 确保 curl 已安装
curl --version

# 如果没有，安装 curl
# macOS: brew install curl
# Ubuntu: sudo apt install curl
```

---

## 更新

### 检查更新

```bash
mini-tmk-agent update --check
```

### 更新到最新版

```bash
mini-tmk-agent update
```

### 手动更新

重新运行安装脚本或下载最新版本覆盖。

---

## 获取帮助

- **GitHub Issues:** https://github.com/luoleixi/MiniTMKAgent/issues
- **阿里云百炼平台:** https://dashscope.console.aliyun.com/

---

## 许可证

MIT License
