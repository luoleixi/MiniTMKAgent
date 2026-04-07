package audio

import (
	"sync"
	"time"
)

// AudioBuffer 线程安全的音频缓冲区
type AudioBuffer struct {
	mu       sync.RWMutex
	data     []int16
	maxSize  int
}

// NewAudioBuffer 创建新的音频缓冲区
func NewAudioBuffer(maxSize int) *AudioBuffer {
	return &AudioBuffer{
		data:    make([]int16, 0, maxSize),
		maxSize: maxSize,
	}
}

// Write 写入数据到缓冲区
func (b *AudioBuffer) Write(samples []int16) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 如果超过最大大小，丢弃旧数据
	if len(b.data)+len(samples) > b.maxSize {
		overflow := len(b.data) + len(samples) - b.maxSize
		if overflow >= len(b.data) {
			b.data = b.data[:0]
		} else {
			b.data = b.data[overflow:]
		}
	}

	b.data = append(b.data, samples...)
}

// Read 从缓冲区读取数据
func (b *AudioBuffer) Read(n int) []int16 {
	b.mu.Lock()
	defer b.mu.Unlock()

	if n > len(b.data) {
		n = len(b.data)
	}

	result := make([]int16, n)
	copy(result, b.data[:n])
	b.data = b.data[n:]

	return result
}

// ReadAll 读取所有数据
func (b *AudioBuffer) ReadAll() []int16 {
	b.mu.Lock()
	defer b.mu.Unlock()

	result := make([]int16, len(b.data))
	copy(result, b.data)
	b.data = b.data[:0]

	return result
}

// Size 返回当前缓冲区大小
func (b *AudioBuffer) Size() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.data)
}

// Clear 清空缓冲区
func (b *AudioBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data = b.data[:0]
}

// SlidingWindowBuffer 滑动窗口缓冲区（用于实时处理）
type SlidingWindowBuffer struct {
	mu          sync.RWMutex
	windowSize  int
	stepSize    int
	buffer      []int16
	onWindow    func([]int16)
	ticker      *time.Ticker
	stopChan    chan struct{}
}

// NewSlidingWindowBuffer 创建滑动窗口缓冲区
func NewSlidingWindowBuffer(windowSize, stepSize int, interval time.Duration, onWindow func([]int16)) *SlidingWindowBuffer {
	return &SlidingWindowBuffer{
		windowSize: windowSize,
		stepSize:   stepSize,
		buffer:     make([]int16, 0, windowSize*2),
		onWindow:   onWindow,
		stopChan:   make(chan struct{}),
		ticker:     time.NewTicker(interval),
	}
}

// Start 开始滑动窗口处理
func (sw *SlidingWindowBuffer) Start() {
	go sw.processLoop()
}

// Stop 停止滑动窗口处理
func (sw *SlidingWindowBuffer) Stop() {
	sw.ticker.Stop()
	close(sw.stopChan)
}

// Write 写入数据
func (sw *SlidingWindowBuffer) Write(samples []int16) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.buffer = append(sw.buffer, samples...)
}

// processLoop 处理循环
func (sw *SlidingWindowBuffer) processLoop() {
	for {
		select {
		case <-sw.ticker.C:
			sw.processWindow()
		case <-sw.stopChan:
			return
		}
	}
}

// processWindow 处理窗口
func (sw *SlidingWindowBuffer) processWindow() {
	sw.mu.Lock()

	if len(sw.buffer) < sw.windowSize {
		sw.mu.Unlock()
		return
	}

	// 提取窗口数据
	window := make([]int16, sw.windowSize)
	copy(window, sw.buffer[:sw.windowSize])

	// 滑动窗口
	if len(sw.buffer) > sw.stepSize {
		sw.buffer = sw.buffer[sw.stepSize:]
	} else {
		sw.buffer = sw.buffer[:0]
	}

	sw.mu.Unlock()

	// 回调处理
	if sw.onWindow != nil {
		sw.onWindow(window)
	}
}
