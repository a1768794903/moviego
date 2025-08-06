package ffmpeg

import (
	"context"
	"fmt"
	"image"
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

// VideoWriter FFmpeg 视频写入器
type VideoWriter struct {
	filename   string
	width      int
	height     int
	fps        float64
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

// VideoWriterOptions 视频写入器选项
type VideoWriterOptions struct {
	Codec   string
	Bitrate string
	FPS     float64
}

// NewVideoWriter 创建新的视频写入器
func NewVideoWriter(filename string, width, height int, options *VideoWriterOptions, processMgr *ProcessManager) *VideoWriter {
	ctx, cancel := context.WithCancel(context.Background())

	// 设置默认选项
	if options == nil {
		options = &VideoWriterOptions{}
	}
	if options.Codec == "" {
		options.Codec = "libx264"
	}
	if options.Bitrate == "" {
		options.Bitrate = "1000k" // 降低比特率以提高兼容性
	}
	if options.FPS == 0 {
		options.FPS = 25.0
	}

	return &VideoWriter{
		filename:   filename,
		width:      width,
		height:     height,
		fps:        options.FPS,
		codec:      options.Codec,
		bitrate:    options.Bitrate,
		processMgr: processMgr,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Open 打开视频写入器
func (vw *VideoWriter) Open() error {
	vw.mutex.Lock()
	defer vw.mutex.Unlock()

	if vw.closed {
		return fmt.Errorf("写入器已关闭")
	}

	// 构建 FFmpeg 命令
	args := []string{
		"-f", "rawvideo",
		"-pix_fmt", "rgb24",
		"-s", fmt.Sprintf("%dx%d", vw.width, vw.height),
		"-r", strconv.FormatFloat(vw.fps, 'f', -1, 64),
		"-i", "-",
		"-c:v", vw.codec,
		"-b:v", vw.bitrate,
		"-preset", "medium", // 编码预设
		"-crf", "23", // 恒定质量因子
		"-pix_fmt", "yuv420p", // 输出像素格式，确保兼容性
		"-threads", "1", // 限制线程数，减少复杂度
		"-loglevel", "verbose", // 显示详细信息用于调试
		"-y", // 覆盖输出文件
		vw.filename,
	}

	// 创建命令
	cmd := exec.CommandContext(vw.ctx, "ffmpeg", args...)

	// 设置stderr到终端，这样可以看到FFmpeg的错误输出
	cmd.Stderr = os.Stderr

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
		ctx:       vw.ctx,
		cancel:    vw.cancel,
		done:      make(chan error, 1),
	}

	// 启动一个 goroutine 来等待进程结束
	go func() {
		err := cmd.Wait()
		process.done <- err
	}()

	vw.process = process
	vw.stdin = stdin

	return nil
}

// WriteFrame 写入一帧
func (vw *VideoWriter) WriteFrame(frame image.Image) error {
	vw.mutex.Lock()
	defer vw.mutex.Unlock()

	if vw.closed {
		return fmt.Errorf("写入器已关闭")
	}

	if vw.process == nil {
		return fmt.Errorf("写入器未打开")
	}

	// 检查帧尺寸
	bounds := frame.Bounds()
	if bounds.Dx() != vw.width || bounds.Dy() != vw.height {
		return fmt.Errorf("帧尺寸不匹配: 期望 %dx%d, 实际 %dx%d",
			vw.width, vw.height, bounds.Dx(), bounds.Dy())
	}

	// 将图像转换为 RGB 字节数组
	pixelData := make([]byte, vw.width*vw.height*3)
	idx := 0

	// 确保从 (0,0) 开始遍历，使用帧的实际尺寸
	for y := 0; y < vw.height; y++ {
		for x := 0; x < vw.width; x++ {
			// 映射到帧的实际坐标
			frameX := bounds.Min.X + x
			frameY := bounds.Min.Y + y

			r, g, b, _ := frame.At(frameX, frameY).RGBA()
			pixelData[idx] = byte(r >> 8)
			pixelData[idx+1] = byte(g >> 8)
			pixelData[idx+2] = byte(b >> 8)
			idx += 3
		}
	}

	// 检查进程是否还在运行
	select {
	case processErr := <-vw.process.done:
		// 进程已经退出
		return fmt.Errorf("FFmpeg进程已退出: %v", processErr)
	default:
		// 进程仍在运行，继续写入
	}

	// 写入数据
	_, err := vw.stdin.Write(pixelData)
	if err != nil {
		// 如果写入失败，检查进程状态
		select {
		case processErr := <-vw.process.done:
			return fmt.Errorf("写入帧数据失败，FFmpeg进程已退出: %v, 写入错误: %w", processErr, err)
		default:
			return fmt.Errorf("写入帧数据失败: %w", err)
		}
	}

	return nil
}

// WriteFrames 批量写入帧
func (vw *VideoWriter) WriteFrames(frames []image.Image) error {
	for i, frame := range frames {
		if err := vw.WriteFrame(frame); err != nil {
			return fmt.Errorf("写入第 %d 帧失败: %w", i, err)
		}
	}
	return nil
}

// Close 关闭写入器
func (vw *VideoWriter) Close() error {
	vw.mutex.Lock()
	defer vw.mutex.Unlock()

	if vw.closed {
		return nil
	}

	vw.closed = true

	// 关闭 stdin
	if vw.stdin != nil {
		vw.stdin.Close()
		vw.stdin = nil
	}

	// 等待进程结束
	if vw.process != nil {
		vw.process.Wait()
		vw.process = nil
	}

	// 取消上下文
	if vw.cancel != nil {
		vw.cancel()
	}

	return nil
}

// IsClosed 检查是否已关闭
func (vw *VideoWriter) IsClosed() bool {
	vw.mutex.RLock()
	defer vw.mutex.RUnlock()
	return vw.closed
}

// GetInfo 获取写入器信息
func (vw *VideoWriter) GetInfo() map[string]interface{} {
	return map[string]interface{}{
		"filename": vw.filename,
		"width":    vw.width,
		"height":   vw.height,
		"fps":      vw.fps,
		"codec":    vw.codec,
		"bitrate":  vw.bitrate,
		"closed":   vw.closed,
	}
}
