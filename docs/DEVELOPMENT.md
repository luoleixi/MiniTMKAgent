# MiniTMK Agent 开发文档

## 项目结构

```
MiniTMKAgent/
├── cmd/                    # CLI 命令入口
│   ├── root.go            # 根命令定义
│   ├── interactive.go     # 交互式菜单模式
│   ├── config.go          # 配置管理命令
│   └── update.go          # 版本更新命令
├── internal/              # 内部实现
│   ├── agent/             # 业务逻辑编排
│   │   └── stream.go      # 流式同传Agent核心
│   ├── audio/             # 音频处理
│   │   ├── player.go      # 音频播放器
│   │   ├── recorder.go    # 音频录制器
│   │   ├── play_queue.go  # 顺序播放队列
│   │   └── decoder.go     # 音频解码
│   ├── client/            # 服务端客户端
│   │   └── relay_client.go # 中继服务客户端
│   ├── config/            # 配置管理
│   │   └── config.go      # 配置文件读写
│   ├── recognizer/        # 语音识别
│   │   ├── interface.go   # 识别器接口
│   │   ├── baiwan_asr.go  # 百炼平台ASR实现
│   │   └── factory.go     # 识别器工厂
│   ├── translator/        # 翻译模块
│   │   ├── interface.go   # 翻译器接口
│   │   ├── baiwan.go      # 百炼平台翻译实现
│   │   └── factory.go     # 翻译器工厂
│   ├── tts/               # 语音合成
│   │   ├── interface.go   # TTS接口
│   │   ├── baiwan.go      # 百炼平台TTS实现
│   │   └── factory.go     # TTS工厂
│   ├── setup/             # 初始化设置
│   │   └── interactive.go # 交互式配置向导
│   └── utils/             # 工具函数
│       ├── logger.go      # 日志工具
│       └── lang.go        # 语言代码处理
└── main.go                # 程序入口
```

---

## 核心模块详解

### 1. Agent 模块 (`internal/agent/stream.go`)

流式同传的核心协调器，管理录音、识别、翻译、TTS全流程。

#### 主要类型

**StreamConfig** - 流式同传配置
```go
type StreamConfig struct {
    SourceLang string // 源语言代码，如 "zh"
    TargetLang string // 目标语言代码，如 "en"
    DirectMode bool   // true=直连百炼平台, false=服务端中转
    ServerURL  string // 服务端地址，如 "http://localhost:8080"
}
```

**StreamAgent** - 流式同传Agent
```go
type StreamAgent struct {
    config     StreamConfig
    recognizer recognizer.Recognizer
    recorder   *audio.Recorder
    vad        *audio.VAD
    translator translator.Translator
    tts        tts.TTS
    player     *audio.Player
    seqGen     *audio.SequenceGenerator
}
```

#### 重要函数

| 函数 | 签名 | 说明 |
|------|------|------|
| NewStreamAgent | `func NewStreamAgent(config StreamConfig) *StreamAgent` | 创建Agent实例 |
| Start | `func (a *StreamAgent) Start(ctx context.Context) error` | 启动翻译流程，根据DirectMode选择模式 |
| startDirect | `func (a *StreamAgent) startDirect(ctx context.Context) error` | 直连百炼平台模式 |
| startWithServer | `func (a *StreamAgent) startWithServer(ctx context.Context) error` | 服务端中转模式 |
| runAudioLoop | `func (a *StreamAgent) runAudioLoop(ctx context.Context) error` | 音频处理主循环（服务端模式） |
| runStreamingAudioLoop | `func (a *StreamAgent) runStreamingAudioLoop(ctx context.Context) error` | 流式音频循环（直连模式） |
| processResults | `func (a *StreamAgent) processResults(ctx context.Context)` | 结果处理协程，从识别器读取结果 |
| handleRecognitionResult | `func (a *StreamAgent) handleRecognitionResult(ctx context.Context, result recognizer.RecognitionResult)` | 处理单次识别结果，调用翻译和TTS |
| playTranslatedText | `func (a *StreamAgent) playTranslatedText(ctx context.Context, seq int, text string)` | 异步TTS合成与播放 |

---

### 2. 语音识别模块 (`internal/recognizer/`)

#### 接口定义 (`interface.go`)

```go
type Recognizer interface {
    Start() error                                    // 启动识别器
    Stop() error                                     // 停止识别器
    SendAudio(audioData []int16) error               // 发送音频数据
    GetResultChan() <-chan RecognitionResult         // 获取识别结果通道
}

type RecognitionResult struct {
    Text       string  // 识别文本
    IsFinal    bool    // 是否是最终结果
    Confidence float64 // 置信度
}
```

#### 百炼ASR实现 (`baiwan_asr.go`)

**BaiwanASR** - 百炼平台实时语音识别器（WebSocket流式）

```go
type BaiwanASR struct {
    apiKey     string                    // 百炼平台API Key
    wsURL      string                    // WebSocket地址
    language   string                    // 识别语言
    sampleRate int                       // 采样率，默认16000
    conn       *websocket.Conn           // WebSocket连接
    resultChan chan RecognitionResult    // 结果通道
    taskID     string                    // 任务ID
}
```

| 函数 | 签名 | 说明 |
|------|------|------|
| newBaiwanASR | `func newBaiwanASR(apiKey, language string) (*BaiwanASR, error)` | 创建ASR实例 |
| Start | `func (r *BaiwanASR) Start() error` | 建立WebSocket连接，发送run-task指令 |
| Stop | `func (r *BaiwanASR) Stop() error` | 发送finish-task，关闭连接 |
| SendAudio | `func (r *BaiwanASR) SendAudio(audioData []int16) error` | 发送音频数据到服务端 |
| GetResultChan | `func (r *BaiwanASR) GetResultChan() <-chan RecognitionResult` | 获取结果通道 |
| sendRunTask | `func (r *BaiwanASR) sendRunTask() error` | 发送启动任务指令 |
| sendFinishTask | `func (r *BaiwanASR) sendFinishTask() error` | 发送结束任务指令 |
| receiveLoop | `func (r *BaiwanASR) receiveLoop()` | 接收消息循环（在独立协程中运行） |
| handleEvent | `func (r *BaiwanASR) handleEvent(message []byte)` | 处理服务端返回的事件 |
| parseResult | `func (r *BaiwanASR) parseResult(event map[string]interface{}) RecognitionResult` | 解析识别结果 |

---

### 3. 翻译模块 (`internal/translator/`)

#### 接口定义 (`interface.go`)

```go
type Translator interface {
    Translate(sourceLang, targetLang, text string) (string, error)
}
```

#### 百炼翻译实现 (`baiwan.go`)

**BaiwanTranslator** - 百炼平台翻译实现（使用qwen-turbo模型）

```go
type BaiwanTranslator struct {
    apiKey string
    client *http.Client
}
```

| 函数 | 签名 | 说明 |
|------|------|------|
| NewBaiwanTranslator | `func NewBaiwanTranslator() (*BaiwanTranslator, error)` | 创建翻译器实例 |
| Translate | `func (t *BaiwanTranslator) Translate(sourceLang, targetLang, text string) (string, error)` | 执行翻译 |
| callBaiwanTranslate | `func (t *BaiwanTranslator) callBaiwanTranslate(sourceLang, targetLang, text string) (string, error)` | 调用百炼API |
| IsConfigured | `func (t *BaiwanTranslator) IsConfigured() bool` | 检查是否已配置 |

---

### 4. TTS模块 (`internal/tts/`)

#### 接口定义 (`interface.go`)

```go
type TTS interface {
    Synthesize(text, voice string) ([]byte, error)                                    // 非流式合成
    SynthesizeStream(text, voice string, onAudioChunk func(chunk []byte)) error      // 流式合成
    Close() error                                                                   // 关闭客户端
}

// 配置常量
const DefaultSampleRate = 22050
const DefaultVoice = "longanyang"
const DefaultModel = "cosyvoice-v3-flash"

// GetVoiceByLanguage 根据语言代码获取音色
func GetVoiceByLanguage(lang string) string
```

#### 百炼TTS实现 (`baiwan.go`)

**BaiwanTTS** - 百炼平台TTS（WebSocket流式API，每请求独立连接）

```go
type BaiwanTTS struct {
    apiKey     string
    wsURL      string
    sampleRate int
    model      string
}

// ttsSession 单次TTS会话（独立连接）
type ttsSession struct {
    apiKey     string
    wsURL      string
    sampleRate int
    model      string
    conn       *websocket.Conn
}
```

| 函数 | 签名 | 说明 |
|------|------|------|
| NewBaiwanTTS | `func NewBaiwanTTS(config *Config) (*BaiwanTTS, error)` | 创建TTS客户端 |
| Synthesize | `func (t *BaiwanTTS) Synthesize(text, voice string) ([]byte, error)` | 非流式合成，返回完整音频 |
| SynthesizeStream | `func (t *BaiwanTTS) SynthesizeStream(text, voice string, onAudioChunk func(chunk []byte)) error` | 流式合成 |
| connect | `func (s *ttsSession) connect() error` | 建立WebSocket连接 |
| synthesize | `func (s *ttsSession) synthesize(text, voice string, onAudioChunk func(chunk []byte)) error` | 执行合成流程 |
| cleanText | `func (t *BaiwanTTS) cleanText(text string) string` | 清理文本（去除emoji等） |

---

### 5. 音频模块 (`internal/audio/`)

#### 播放器 (`player.go`)

**Player** - 基于malgo的音频播放器

```go
type Player struct {
    ctx        *malgo.AllocatedContext
    device     *malgo.Device
    sampleRate int
    channels   int
    isRunning  bool
    dataChan   chan []int16
}
```

| 函数 | 签名 | 说明 |
|------|------|------|
| NewPlayer | `func NewPlayer(sampleRate int) (*Player, error)` | 创建播放器 |
| Start | `func (p *Player) Start() error` | 初始化malgo上下文，启动播放设备 |
| Stop | `func (p *Player) Stop() error` | 停止播放，释放资源 |
| Write | `func (p *Player) Write(data []int16) error` | 写入音频数据（分块写入避免阻塞） |
| onData | `func (p *Player) onData(pOutputSamples, pInputSamples []byte, frameCount uint32)` | 播放回调函数 |

#### 录音器 (`recorder.go`)

**Recorder** - 基于malgo的音频录制器

```go
type Recorder struct {
    ctx        *malgo.AllocatedContext
    device     *malgo.Device
    sampleRate int
    channels   int
    isRunning  bool
    dataChan   chan []int16
}
```

| 函数 | 签名 | 说明 |
|------|------|------|
| NewRecorder | `func NewRecorder(sampleRate int) (*Recorder, error)` | 创建录音器 |
| Start | `func (r *Recorder) Start() error` | 启动录音设备 |
| Stop | `func (r *Recorder) Stop() error` | 停止录音，关闭通道 |
| Read | `func (r *Recorder) Read() ([]int16, bool)` | 读取音频数据（阻塞） |
| GetDataChan | `func (r *Recorder) GetDataChan() <-chan []int16` | 获取数据通道 |
| onData | `func (r *Recorder) onData(pOutputSample, pInputSamples []byte, frameCount uint32)` | 录音回调函数 |

#### 播放队列 (`play_queue.go`)

**PlayQueue** - 确保按序列号顺序播放的队列（全局单例）

```go
type PlayQueue struct {
    queue       chan QueuedAudio
    pending     map[int][]byte    // 乱序到达的音频缓存
    nextSeq     int               // 下一个期望的序列号
    mu          sync.Mutex
    isPlaying   bool
    stopChan    chan struct{}
    sequenceGen *SequenceGenerator
}

type QueuedAudio struct {
    Sequence  int
    Data      []byte
    Timestamp time.Time
}

// SequenceGenerator 序列号生成器
type SequenceGenerator struct {
    mu      sync.Mutex
    current int
}
```

| 函数 | 签名 | 说明 |
|------|------|------|
| GetPlayQueue | `func GetPlayQueue() *PlayQueue` | 获取全局播放队列单例 |
| Enqueue | `func (pq *PlayQueue) Enqueue(seq int, data []byte)` | 将音频加入队列（带序列号） |
| EnqueueWithoutSeq | `func (pq *PlayQueue) EnqueueWithoutSeq(data []byte) int` | 自动分配序列号并加入队列 |
| NextSequence | `func (pq *PlayQueue) NextSequence() int` | 获取下一个序列号 |
| playLoop | `func (pq *PlayQueue) playLoop()` | 后台播放协程 |
| handleAudio | `func (pq *PlayQueue) handleAudio(audio QueuedAudio)` | 处理音频，确保顺序播放 |
| playAndContinue | `func (pq *PlayQueue) playAndContinue(data []byte)` | 播放音频并继续检查后续 |
| playPendingSequence | `func (pq *PlayQueue) playPendingSequence()` | 按顺序播放待处理的音频 |
| Reset | `func (pq *PlayQueue) Reset()` | 重置队列状态 |

---

### 6. 服务端客户端 (`internal/client/relay_client.go`)

**RelayClient** - 中继服务器HTTP客户端

```go
type RelayClient struct {
    baseURL string
    client  *http.Client
}

type RelayConfig struct {
    BaseURL string
}
```

| 函数 | 签名 | 说明 |
|------|------|------|
| NewRelayClient | `func NewRelayClient(config *RelayConfig) *RelayClient` | 创建中继客户端 |
| CheckHealth | `func (c *RelayClient) CheckHealth() (map[string]interface{}, error)` | 检查服务端健康状态 |
| GetASRToken | `func (c *RelayClient) GetASRToken() (*ASRTokenResponse, error)` | 获取ASR临时Token |
| NewRelayTranslator | `func (c *RelayClient) NewRelayTranslator() translator.Translator` | 创建中继翻译器 |
| NewRelayRecognizer | `func (c *RelayClient) NewRelayRecognizer(language string) recognizer.Recognizer` | 创建中继识别器 |
| NewRelayTTS | `func (c *RelayClient) NewRelayTTS() tts.TTS` | 创建中继TTS |

---

### 7. 配置模块 (`internal/config/config.go`)

**Config** - 应用配置（单例模式）

```go
type Config struct {
    Mode      string       // "server" 或 "direct"
    ServerURL string       // 服务端地址
    Baiwan    BaiwanConfig // 百炼平台配置
}

type BaiwanConfig struct {
    APIKey string
}
```

| 函数 | 签名 | 说明 |
|------|------|------|
| GetInstance | `func GetInstance() *Config` | 获取配置单例（延迟加载） |
| Save | `func (c *Config) Save() error` | 保存配置到文件 |
| GetConfigPath | `func GetConfigPath() (string, error)` | 获取配置文件路径 |
| SetBaiwanAPIKey | `func (c *Config) SetBaiwanAPIKey(apiKey string) error` | 设置百炼API Key |
| GetBaiwanAPIKey | `func (c *Config) GetBaiwanAPIKey() (apiKey string, ok bool)` | 获取百炼API Key |
| IsConfigured | `func (c *Config) IsConfigured() bool` | 检查是否已配置 |
| SetMode | `func (c *Config) SetMode(mode, serverURL string) error` | 设置运行模式 |
| GetMode | `func (c *Config) GetMode() (mode, serverURL string)` | 获取运行模式 |
| MaskKey | `func MaskKey(key string) string` | 隐藏密钥中间部分 |

---

### 8. 命令模块 (`cmd/`)

#### 根命令 (`root.go`)

| 函数/变量 | 说明 |
|-----------|------|
| rootCmd | 根命令定义，默认进入interactive模式 |
| Execute | 执行根命令 |
| runStreamDefault | 使用默认设置启动流式翻译 |

#### 交互式模式 (`interactive.go`)

| 函数 | 说明 |
|------|------|
| runInteractive | 运行交互式菜单模式 |
| showInteractiveMenu | 显示主菜单 |
| startInteractiveStreaming | 启动流式翻译（交互式） |
| selectLanguage | 语言选择交互 |

#### 配置命令 (`config.go`)

| 命令 | 说明 |
|------|------|
| config show | 显示当前配置 |
| config set-baiwan-key | 设置百炼平台API Key |
| config set-mode | 设置运行模式 |
| config reset | 清除所有配置 |

---

## 数据流图

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Recorder  │────▶│     VAD     │────▶│  Recognizer │
│  (录音设备)  │     │ (语音检测)   │     │  (语音识别)  │
└─────────────┘     └─────────────┘     └──────┬──────┘
                                               │
                                               ▼
                                        ┌─────────────┐
                                        │  Translator │
                                        │   (翻译)     │
                                        └──────┬──────┘
                                               │
                                               ▼
                                        ┌─────────────┐
                                        │     TTS     │
                                        │  (语音合成)  │
                                        └──────┬──────┘
                                               │
                                               ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Speaker   │◀────│  PlayQueue  │◀────│   Player    │
│   (扬声器)   │     │ (顺序队列)   │     │  (播放设备)  │
└─────────────┘     └─────────────┘     └─────────────┘
```

---

## 添加新功能

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

### 添加新的识别器/翻译器/TTS实现

1. 实现对应的接口
2. 在 factory.go 中添加创建逻辑
3. 通过配置或命令行参数切换
