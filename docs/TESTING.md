# MiniTMK Agent 测试指南

## 测试结构

```
MiniTMKAgent/
├── internal/
│   ├── audio/
│   │   ├── vad_test.go         # VAD 测试
│   │   └── play_queue_test.go  # 播放队列测试
│   ├── utils/
│   │   └── lang_test.go        # 语言工具测试
│   ├── config/
│   │   └── config_test.go      # 配置测试
│   └── tts/
│       └── tts_test.go         # TTS 测试
└── tests/
    └── integration_test.go     # 集成测试
```

## 运行测试

### 运行所有测试

```bash
go test ./...
```

### 运行特定包的测试

```bash
# 测试 utils 包
go test ./internal/utils/...

# 测试 audio 包
go test ./internal/audio/...

# 测试 config 包
go test ./internal/config/...
```

### 运行特定测试函数

```bash
# 运行单个测试
go test -run TestValidateLanguagePair ./internal/utils/...

# 运行匹配的测试
go test -run "TestLang" ./internal/utils/...
```

### 详细输出

```bash
go test -v ./...
```

### 覆盖率报告

```bash
# 生成覆盖率报告
go test -cover ./...

# 生成详细覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### 基准测试

```bash
# 运行基准测试
go test -bench=. ./...

# 运行基准测试并显示内存分配
go test -bench=. -benchmem ./...
```

## 测试类型

### 1. 单元测试

测试单个函数或方法：

```go
func TestValidateLanguagePair(t *testing.T) {
    err := ValidateLanguagePair("zh", "en")
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
}
```

### 2. 表驱动测试

使用表格形式测试多组数据：

```go
func TestIsValidLangCode(t *testing.T) {
    tests := []struct {
        code string
        want bool
    }{
        {"zh", true},
        {"en", true},
        {"xx", false},
    }

    for _, tt := range tests {
        t.Run(tt.code, func(t *testing.T) {
            if got := IsValidLangCode(tt.code); got != tt.want {
                t.Errorf("IsValidLangCode(%s) = %v, want %v", tt.code, got, tt.want)
            }
        })
    }
}
```

### 3. 并发测试

测试并发安全性：

```go
func TestSequenceGenerator_Concurrent(t *testing.T) {
    gen := NewSequenceGenerator()

    done := make(chan bool, 10)
    for i := 0; i < 10; i++ {
        go func() {
            for j := 0; j < 10; j++ {
                gen.Next()
            }
            done <- true
        }()
    }

    for i := 0; i < 10; i++ {
        <-done
    }

    if got := gen.Current(); got != 100 {
        t.Errorf("Expected 100, got %d", got)
    }
}
```

### 4. Mock 测试

对于外部依赖，使用接口进行 Mock：

```go
// 定义接口
type TTS interface {
    Synthesize(text, voice string) ([]byte, error)
}

// Mock 实现
type MockTTS struct {
    SynthesizeFunc func(text, voice string) ([]byte, error)
}

func (m *MockTTS) Synthesize(text, voice string) ([]byte, error) {
    return m.SynthesizeFunc(text, voice)
}
```

## 测试规范

### 命名规范

- 测试函数: `TestXxx` 或 `Test_Xxx_Yyy`
- 子测试: `t.Run("descriptive name", ...)`
- 基准测试: `BenchmarkXxx`
- 示例测试: `ExampleXxx`

### 测试结构

```go
func TestFeature(t *testing.T) {
    // 准备
    setup()
    defer teardown()

    // 执行
    result, err := DoSomething()

    // 验证
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

### 常用断言模式

```go
// 错误检查
if err != nil {
    t.Errorf("Function() error = %v", err)
}

// 期望值检查
if got != want {
    t.Errorf("Function() = %v, want %v", got, want)
}

// 错误预期
if (err != nil) != wantErr {
    t.Errorf("Function() error = %v, wantErr %v", err, wantErr)
}

// Fatal vs Error
// t.Fatal() - 立即停止当前测试
// t.Errorf() - 记录错误但继续测试
```

## 添加新测试

1. 在对应包中创建 `xxx_test.go` 文件
2. 编写测试函数
3. 运行测试确保通过
4. 提交代码

## 持续集成

### GitHub Actions 示例

```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go test -v -race -cover ./...
```

## 注意事项

1. **不要测试外部服务**：使用 Mock 代替真实的 API 调用
2. **测试应该是独立的**：每个测试可以单独运行
3. **清理测试数据**：使用 `t.TempDir()` 或 `defer` 清理
4. **避免全局状态**：测试间不应相互影响
5. **测试覆盖率目标**：核心模块覆盖率 > 80%

## 调试测试

```bash
# 打印详细信息
go test -v -run TestXxx

# 使用 delve 调试
dlv test ./internal/utils -- -test.run TestXxx
```
