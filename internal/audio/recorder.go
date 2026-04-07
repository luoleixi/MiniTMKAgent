package audio

import (
	"fmt"
	"strings"

	"github.com/gen2brain/malgo"
)

// Recorder 音频录制器
type Recorder struct {
	ctx        *malgo.AllocatedContext
	device     *malgo.Device
	sampleRate int
	channels   int
	isRunning  bool
	dataChan   chan []int16
}

// NewRecorder 创建新的音频录制器
func NewRecorder(sampleRate int) (*Recorder, error) {
	return &Recorder{
		sampleRate: sampleRate,
		channels:   1, // 单声道
		dataChan:   make(chan []int16, 100),
	}, nil
}

// Start 开始录制
func (r *Recorder) Start() error {
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
	r.ctx = ctx

	// 配置设备
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = uint32(r.channels)
	deviceConfig.SampleRate = uint32(r.sampleRate)
	// 注意：Alsa.NoMMap 仅用于 Linux，Windows 忽略此设置

	// 创建设备
	device, err := malgo.InitDevice(ctx.Context, deviceConfig, malgo.DeviceCallbacks{
		Data: r.onData,
	})
	if err != nil {
		ctx.Uninit()
		return fmt.Errorf("创建音频设备失败: %w", err)
	}

	r.device = device

	// 启动设备
	if err := device.Start(); err != nil {
		device.Uninit()
		ctx.Uninit()
		return fmt.Errorf("启动音频设备失败: %w", err)
	}

	r.isRunning = true
	fmt.Printf("[音频] 录音已启动: %d Hz, %d 通道\n", r.sampleRate, r.channels)
	return nil
}

// onData 音频数据回调
func (r *Recorder) onData(pOutputSample, pInputSamples []byte, frameCount uint32) {
	// 检查是否还在运行
	if !r.isRunning {
		return
	}

	// 将字节数据转换为 int16
	samples := make([]int16, frameCount)
	for i := uint32(0); i < frameCount; i++ {
		samples[i] = int16(pInputSamples[i*2]) | int16(pInputSamples[i*2+1])<<8
	}

	select {
	case r.dataChan <- samples:
	default:
		// channel满，丢弃数据
	}
}

// Read 读取音频数据
func (r *Recorder) Read() ([]int16, bool) {
	data, ok := <-r.dataChan
	return data, ok
}

// Stop 停止录制
func (r *Recorder) Stop() error {
	if !r.isRunning {
		return nil
	}

	r.isRunning = false

	// 先关闭通道，防止发送端阻塞
	close(r.dataChan)

	if r.device != nil {
		r.device.Stop()
		r.device.Uninit()
	}

	if r.ctx != nil {
		r.ctx.Uninit()
	}

	return nil
}

// GetDataChan 获取数据通道
func (r *Recorder) GetDataChan() <-chan []int16 {
	return r.dataChan
}

// isDeviceInfo 判断是否为需要显示的设备关键信息
func isDeviceInfo(message string) bool {
	// 只显示设备名称行（包含括号但不包含详细的参数信息）
	if strings.Contains(message, "(") && strings.Contains(message, ")") {
		// 过滤掉包含详细参数的输出行
		if strings.Contains(message, "Buffer Size") ||
			strings.Contains(message, "Format:") ||
			strings.Contains(message, "Channels:") ||
			strings.Contains(message, "Sample Rate:") ||
			strings.Contains(message, "Conversion:") ||
			strings.Contains(message, "Trying") {
			return false
		}
		return true
	}

	// 显示错误信息
	if strings.Contains(message, "错误") || strings.Contains(message, "Error") || strings.Contains(message, "Failed") {
		return true
	}

	// 过滤掉其他调试信息
	return false
}
