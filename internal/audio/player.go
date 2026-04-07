package audio

import (
	"fmt"
	"time"

	"github.com/gen2brain/malgo"
)

// Player 音频播放器
type Player struct {
	ctx        *malgo.AllocatedContext
	device     *malgo.Device
	sampleRate int
	channels   int
	isRunning  bool
	dataChan   chan []int16
}

// NewPlayer 创建新的音频播放器
func NewPlayer(sampleRate int) (*Player, error) {
	return &Player{
		sampleRate: sampleRate,
		channels:   1,
		dataChan:   make(chan []int16, 500), // 💡 增大缓冲区避免播放中断
	}, nil
}

// Start 开始播放
func (p *Player) Start() error {
	// 初始化 malgo 上下文（过滤详细日志，只显示设备信息）
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		// 只显示设备名称等关键信息，过滤掉详细的调试信息
		if isDeviceInfo(message) {
			fmt.Printf("[音频] %s\n", message)
		}
	})
	if err != nil {
		return fmt.Errorf("初始化音频上下文失败: %w", err)
	}
	p.ctx = ctx

	// 配置设备
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)
	deviceConfig.Playback.Format = malgo.FormatS16
	deviceConfig.Playback.Channels = uint32(p.channels)
	deviceConfig.SampleRate = uint32(p.sampleRate)
	deviceConfig.Alsa.NoMMap = 1

	// 创建设备
	device, err := malgo.InitDevice(ctx.Context, deviceConfig, malgo.DeviceCallbacks{
		Data: p.onData,
	})
	if err != nil {
		ctx.Uninit()
		return fmt.Errorf("创建音频设备失败: %w", err)
	}

	p.device = device

	// 启动设备
	if err := device.Start(); err != nil {
		device.Uninit()
		ctx.Uninit()
		return fmt.Errorf("启动音频设备失败: %w", err)
	}

	p.isRunning = true
	return nil
}

// onData 播放数据回调
func (p *Player) onData(pOutputSamples, pInputSamples []byte, frameCount uint32) {
	select {
	case data := <-p.dataChan:
		// 复制数据到输出缓冲区
		n := int(frameCount)
		if len(data) < n {
			n = len(data)
		}
		for i := 0; i < n; i++ {
			pOutputSamples[i*2] = byte(data[i])
			pOutputSamples[i*2+1] = byte(data[i] >> 8)
		}
		// 剩余部分填充静音
		for i := n; i < int(frameCount); i++ {
			pOutputSamples[i*2] = 0
			pOutputSamples[i*2+1] = 0
		}
	default:
		// 没有数据，播放静音
		for i := uint32(0); i < frameCount*2; i++ {
			pOutputSamples[i] = 0
		}
	}
}

// Write 写入音频数据
func (p *Player) Write(data []int16) error {
	if !p.isRunning {
		return fmt.Errorf("播放器未启动")
	}

	// 💡 将大数据分块写入，避免阻塞和电音
	const chunkSize = 3200 // 200ms 音频数据 @ 16kHz

	for len(data) > 0 {
		size := chunkSize
		if len(data) < size {
			size = len(data)
		}

		select {
		case p.dataChan <- data[:size]:
			data = data[size:]
		default:
			// 缓冲区满时短暂等待
			if len(p.dataChan) >= cap(p.dataChan)-1 {
				time.Sleep(10 * time.Millisecond)
			}
		}
	}

	return nil
}

// Stop 停止播放
func (p *Player) Stop() error {
	if !p.isRunning {
		return nil
	}

	p.isRunning = false

	if p.device != nil {
		p.device.Stop()
		p.device.Uninit()
	}

	if p.ctx != nil {
		p.ctx.Uninit()
	}

	close(p.dataChan)
	return nil
}
