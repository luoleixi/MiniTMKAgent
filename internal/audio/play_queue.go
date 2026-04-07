package audio

import (
	"sync"
)

// PlayQueue 音频播放队列，确保按顺序播放
type PlayQueue struct {
	queue     chan []byte
	isPlaying bool
	mu        sync.Mutex
	stopChan  chan struct{}
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
			queue:    make(chan []byte, 100),
			stopChan: make(chan struct{}),
		}
		// 启动后台播放协程
		go globalQueue.playLoop()
	})
	return globalQueue
}

// Enqueue 将音频数据加入播放队列
func (pq *PlayQueue) Enqueue(data []byte) {
	if len(data) == 0 {
		return
	}
	// 复制数据避免被修改
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	pq.queue <- dataCopy
}

// playLoop 后台播放循环，按顺序播放队列中的音频
func (pq *PlayQueue) playLoop() {
	for {
		select {
		case data := <-pq.queue:
			if len(data) > 0 {
				pq.playSync(data)
			}
		case <-pq.stopChan:
			return
		}
	}
}

// playSync 同步播放单条音频
func (pq *PlayQueue) playSync(data []byte) {
	pq.mu.Lock()
	pq.isPlaying = true
	pq.mu.Unlock()

	// 直接调用内部播放器，不经过 PlayMP3Data 避免循环
	player, err := GetGoPlayer()
	if err == nil {
		// 调用纯 Go 播放器直接播放
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
