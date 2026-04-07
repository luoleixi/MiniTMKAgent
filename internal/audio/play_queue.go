package audio

import (
	"sync"
	"time"
)

// QueuedAudio 带序列号的音频数据
type QueuedAudio struct {
	Sequence  int
	Data      []byte
	Timestamp time.Time
}

// PlayQueue 音频播放队列，确保按序列号顺序播放
type PlayQueue struct {
	queue       chan QueuedAudio
	pending     map[int][]byte
	nextSeq     int
	mu          sync.Mutex
	isPlaying   bool
	stopChan    chan struct{}
	sequenceGen *SequenceGenerator
}

var (
	// 全局播放队列单例
	globalQueue     *PlayQueue
	globalQueueOnce sync.Once
)

// GetPlayQueue 获取全局播放队列
func GetPlayQueue() *PlayQueue {
	globalQueueOnce.Do(func() {
		globalQueue = &PlayQueue{
			queue:       make(chan QueuedAudio, 100),
			pending:     make(map[int][]byte),
			nextSeq:     1,
			stopChan:    make(chan struct{}),
			sequenceGen: NewSequenceGenerator(),
		}
		// 启动后台播放协程
		go globalQueue.playLoop()
	})
	return globalQueue
}

// NextSequence 获取下一个序列号（线程安全）
func (pq *PlayQueue) NextSequence() int {
	return pq.sequenceGen.Next()
}

// CurrentSequence 获取当前序列号
func (pq *PlayQueue) CurrentSequence() int {
	return pq.sequenceGen.Current()
}

// Enqueue 将音频数据加入播放队列（带序列号）
func (pq *PlayQueue) Enqueue(seq int, data []byte) {
	if len(data) == 0 {
		return
	}
	// 复制数据避免被修改
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	select {
	case pq.queue <- QueuedAudio{Sequence: seq, Data: dataCopy, Timestamp: time.Now()}:
	default:
		// 队列满，直接丢弃（避免阻塞）
	}
}

// EnqueueWithoutSeq 将音频数据加入队列（自动分配序列号）
func (pq *PlayQueue) EnqueueWithoutSeq(data []byte) int {
	seq := pq.NextSequence()
	pq.Enqueue(seq, data)
	return seq
}

// playLoop 后台播放循环，按序列号顺序播放
func (pq *PlayQueue) playLoop() {
	for {
		select {
		case audio := <-pq.queue:
			pq.handleAudio(audio)
		case <-pq.stopChan:
			return
		}
	}
}

// handleAudio 处理音频，确保按序列号顺序播放
func (pq *PlayQueue) handleAudio(audio QueuedAudio) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	// 如果是期望的下一个序列号，立即播放
	if audio.Sequence == pq.nextSeq {
		pq.playAndContinue(audio.Data)
		return
	}

	// 否则存入待处理队列
	pq.pending[audio.Sequence] = audio.Data

	// 检查是否有可以连续播放的音频
	pq.playPendingSequence()
}

// playAndContinue 播放音频并继续检查后续序列
func (pq *PlayQueue) playAndContinue(data []byte) {
	pq.mu.Unlock()
	pq.playSync(data)
	pq.mu.Lock()

	pq.nextSeq++

	// 检查是否有后续的音频可以播放
	pq.playPendingSequence()
}

// playPendingSequence 按顺序播放待处理的音频
func (pq *PlayQueue) playPendingSequence() {
	for {
		data, ok := pq.pending[pq.nextSeq]
		if !ok {
			break
		}
		delete(pq.pending, pq.nextSeq)
		pq.mu.Unlock()
		pq.playSync(data)
		pq.mu.Lock()
		pq.nextSeq++
	}
}

// playSync 同步播放单条音频
func (pq *PlayQueue) playSync(data []byte) {
	pq.mu.Lock()
	pq.isPlaying = true
	pq.mu.Unlock()

	// 直接调用内部播放器
	player, err := GetGoPlayer()
	if err == nil {
		_ = player.PlayMP3Data(data)
	}

	pq.mu.Lock()
	pq.isPlaying = false
	pq.mu.Unlock()
}

// IsPlaying 检查是否正在播放
func (pq *PlayQueue) IsPlaying() bool {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	return pq.isPlaying
}

// Stop 停止播放队列
func (pq *PlayQueue) Stop() {
	close(pq.stopChan)
}

// QueueSize 返回队列中待播放的音频数量
func (pq *PlayQueue) QueueSize() int {
	return len(pq.queue)
}

// PendingSize 返回待处理（乱序到达）的音频数量
func (pq *PlayQueue) PendingSize() int {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	return len(pq.pending)
}

// Reset 重置播放队列状态
func (pq *PlayQueue) Reset() {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	pq.nextSeq = 1
	pq.pending = make(map[int][]byte)
	pq.sequenceGen.Reset()
}

// SequenceGenerator 序列号生成器
type SequenceGenerator struct {
	mu      sync.Mutex
	current int
}

// NewSequenceGenerator 创建序列号生成器
func NewSequenceGenerator() *SequenceGenerator {
	return &SequenceGenerator{current: 0}
}

// Next 获取下一个序列号
func (sg *SequenceGenerator) Next() int {
	sg.mu.Lock()
	defer sg.mu.Unlock()
	sg.current++
	return sg.current
}

// Current 获取当前序列号
func (sg *SequenceGenerator) Current() int {
	sg.mu.Lock()
	defer sg.mu.Unlock()
	return sg.current
}

// Reset 重置序列号
func (sg *SequenceGenerator) Reset() {
	sg.mu.Lock()
	defer sg.mu.Unlock()
	sg.current = 0
}
