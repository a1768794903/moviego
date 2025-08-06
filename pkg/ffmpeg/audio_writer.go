package ffmpeg

import (
	"context"
	"fmt"
	"io"
	"math"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

// AudioWriter FFmpeg 音频写入器
type AudioWriter struct {
	filename   string
	sampleRate int
	channels   int
	codec      string
	bitrate    string
	processMgr *ProcessManager
	process    *ManagedProcess
	ctx        context.Context
	cancel     context.CancelFunc
	closed     bool
	mutex      sync.RWMutex
	stdin      io.WriteCloser
}

// AudioWriterOptions 音频写入器选项
type AudioWriterOptions struct {
	Codec      string
	Bitrate    string
	SampleRate int
	Channels   int
}

// NewAudioWriter 创建新的音频写入器
func NewAudioWriter(filename string, options *AudioWriterOptions, processMgr *ProcessManager) *AudioWriter {
	ctx, cancel := context.WithCancel(context.Background())

	// 设置默认选项
	if options == nil {
		options = &AudioWriterOptions{}
	}
	if options.Codec == "" {
		options.Codec = "aac"
	}
	if options.Bitrate == "" {
		options.Bitrate = "128k"
	}
	if options.SampleRate == 0 {
		options.SampleRate = 44100
	}
	if options.Channels == 0 {
		options.Channels = 2
	}

	return &AudioWriter{
		filename:   filename,
		sampleRate: options.SampleRate,
		channels:   options.Channels,
		codec:      options.Codec,
		bitrate:    options.Bitrate,
		processMgr: processMgr,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Open 打开音频写入器
func (aw *AudioWriter) Open() error {
	aw.mutex.Lock()
	defer aw.mutex.Unlock()

	if aw.closed {
		return fmt.Errorf("写入器已关闭")
	}

	// 构建 FFmpeg 命令
	args := []string{
		//"-f", "f32le", // 输入格式：32位浮点
		"-ar", strconv.Itoa(aw.sampleRate), // 采样率
		"-ac", strconv.Itoa(aw.channels), // 声道数
		"-i", "-", // 从stdin读取
		"-c:a", aw.codec, // 音频编码器
		"-b:a", aw.bitrate, // 音频比特率
		"-y",        // 覆盖输出文件
		aw.filename, // 输出文件
	}

	// 创建命令
	cmd := exec.CommandContext(aw.ctx, "ffmpeg", args...)

	// 在启动进程之前设置输入管道
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("设置输入管道失败: %w", err)
	}

	// 启动进程
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 FFmpeg 失败: %w", err)
	}

	// 创建进程包装器
	process := &ManagedProcess{
		cmd:       cmd,
		pid:       cmd.Process.Pid,
		startTime: time.Now(),
		ctx:       aw.ctx,
		cancel:    aw.cancel,
		done:      make(chan error, 1),
	}

	// 注册到进程管理器
	aw.processMgr.mutex.Lock()
	aw.processMgr.processes[process.pid] = process
	aw.processMgr.mutex.Unlock()

	// 启动一个 goroutine 来等待进程结束
	go func() {
		err := cmd.Wait()
		process.done <- err

		// 从管理器中移除
		aw.processMgr.mutex.Lock()
		delete(aw.processMgr.processes, process.pid)
		aw.processMgr.mutex.Unlock()
	}()

	aw.process = process
	aw.stdin = stdin

	return nil
}

// WriteSamples 写入音频样本
func (aw *AudioWriter) WriteSamples(samples []float64) error {
	aw.mutex.Lock()
	defer aw.mutex.Unlock()

	if aw.closed {
		return fmt.Errorf("写入器已关闭")
	}

	if aw.process == nil {
		return fmt.Errorf("写入器未打开")
	}

	// 将浮点数转换为字节数组
	audioData := make([]byte, len(samples)*4)
	for i, sample := range samples {
		// 将浮点数转换为32位浮点数（IEEE 754格式）
		value := math.Float32bits(float32(sample))
		offset := i * 4

		// 小端序写入
		audioData[offset] = byte(value)
		audioData[offset+1] = byte(value >> 8)
		audioData[offset+2] = byte(value >> 16)
		audioData[offset+3] = byte(value >> 24)
	}

	// 写入数据
	_, err := aw.stdin.Write(audioData)
	if err != nil {
		return fmt.Errorf("写入音频数据失败: %w", err)
	}

	return nil
}

// WriteAudioFrame 写入音频帧
func (aw *AudioWriter) WriteAudioFrame(frame []float64) error {
	return aw.WriteSamples(frame)
}

// Close 关闭写入器
func (aw *AudioWriter) Close() error {
	aw.mutex.Lock()
	defer aw.mutex.Unlock()

	if aw.closed {
		return nil
	}

	aw.closed = true

	// 关闭 stdin
	if aw.stdin != nil {
		aw.stdin.Close()
		aw.stdin = nil
	}

	// 等待进程结束
	if aw.process != nil {
		aw.process.Wait()
		aw.process = nil
	}

	// 取消上下文
	if aw.cancel != nil {
		aw.cancel()
	}

	return nil
}

// IsClosed 检查是否已关闭
func (aw *AudioWriter) IsClosed() bool {
	aw.mutex.RLock()
	defer aw.mutex.RUnlock()
	return aw.closed
}

// GetInfo 获取写入器信息
func (aw *AudioWriter) GetInfo() map[string]interface{} {
	return map[string]interface{}{
		"filename":   aw.filename,
		"sampleRate": aw.sampleRate,
		"channels":   aw.channels,
		"codec":      aw.codec,
		"bitrate":    aw.bitrate,
		"closed":     aw.closed,
	}
}
