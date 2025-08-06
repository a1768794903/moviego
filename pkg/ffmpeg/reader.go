package ffmpeg

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// VideoInfo 视频信息
type VideoInfo struct {
	Duration        float64 `json:"duration"`
	Width           int     `json:"width"`
	Height          int     `json:"height"`
	FPS             float64 `json:"fps"`
	BitRate         string  `json:"bit_rate"`
	Codec           string  `json:"codec_name"`
	HasAudio        bool    `json:"has_audio"`
	AudioCodec      string  `json:"audio_codec"`
	AudioSampleRate int     `json:"audio_sample_rate"`
	AudioChannels   int     `json:"audio_channels"`
}

// VideoReader FFmpeg 视频读取器
type VideoReader struct {
	filename   string
	info       *VideoInfo
	processMgr *ProcessManager
	process    *ManagedProcess
	ctx        context.Context
	cancel     context.CancelFunc
	closed     bool
	mutex      sync.RWMutex
}

// NewVideoReader 创建新的视频读取器
func NewVideoReader(filename string, processMgr *ProcessManager) *VideoReader {
	ctx, cancel := context.WithCancel(context.Background())
	return &VideoReader{
		filename:   filename,
		processMgr: processMgr,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Open 打开视频文件并获取信息
func (vr *VideoReader) Open() error {
	vr.mutex.Lock()
	defer vr.mutex.Unlock()

	if vr.closed {
		return fmt.Errorf("读取器已关闭")
	}

	// 检查文件是否存在
	if _, err := os.Stat(vr.filename); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", vr.filename)
	}

	// 获取视频信息
	info, err := vr.getVideoInfo()
	if err != nil {
		return fmt.Errorf("获取视频信息失败: %w", err)
	}

	vr.info = info
	return nil
}

// getVideoInfo 获取视频信息
func (vr *VideoReader) getVideoInfo() (*VideoInfo, error) {
	args := []string{
		"-i", vr.filename,
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
	}

	cmd := exec.Command("ffprobe", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe 执行失败: %w", err)
	}

	var result struct {
		Format struct {
			Duration string `json:"duration"`
			BitRate  string `json:"bit_rate"`
		} `json:"format"`
		Streams []struct {
			CodecType  string `json:"codec_type"`
			CodecName  string `json:"codec_name"`
			Width      int    `json:"width"`
			Height     int    `json:"height"`
			RFrameRate string `json:"r_frame_rate"`
			SampleRate string `json:"sample_rate"`
			Channels   int    `json:"channels"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	info := &VideoInfo{}

	// 解析时长
	if result.Format.Duration != "" {
		if duration, err := strconv.ParseFloat(result.Format.Duration, 64); err == nil {
			info.Duration = duration
		}
	}

	info.BitRate = result.Format.BitRate

	// 解析视频流
	for _, stream := range result.Streams {
		if stream.CodecType == "video" {
			info.Width = stream.Width
			info.Height = stream.Height
			info.Codec = stream.CodecName

			// 解析帧率
			if stream.RFrameRate != "" {
				parts := strings.Split(stream.RFrameRate, "/")
				if len(parts) == 2 {
					if num, err := strconv.ParseFloat(parts[0], 64); err == nil {
						if den, err := strconv.ParseFloat(parts[1], 64); err == nil && den != 0 {
							info.FPS = num / den
						}
					}
				}
			}
		} else if stream.CodecType == "audio" {
			info.HasAudio = true
			info.AudioCodec = stream.CodecName
			info.AudioChannels = stream.Channels

			if stream.SampleRate != "" {
				if sampleRate, err := strconv.Atoi(stream.SampleRate); err == nil {
					info.AudioSampleRate = sampleRate
				}
			}
		}
	}

	return info, nil
}

// GetFrame 获取指定时间的帧
func (vr *VideoReader) GetFrame(t time.Duration) (image.Image, error) {
	vr.mutex.RLock()
	defer vr.mutex.RUnlock()

	if vr.closed {
		return nil, fmt.Errorf("读取器已关闭")
	}

	if vr.info == nil {
		return nil, fmt.Errorf("视频未打开")
	}

	// 计算时间戳
	timestamp := t.Seconds()
	if timestamp > vr.info.Duration {
		return nil, fmt.Errorf("时间超出视频长度")
	}

	// 启动 FFmpeg 进程读取帧
	args := []string{
		"-ss", fmt.Sprintf("%.3f", timestamp),
		"-i", vr.filename,
		"-vframes", "1",
		"-f", "image2pipe",
		"-pix_fmt", "rgb24",
		"-vcodec", "rawvideo",
		"-",
	}

	// 创建命令
	cmd := exec.CommandContext(vr.ctx, "ffmpeg", args...)

	// 在启动进程之前设置输出管道
	output, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("设置输出管道失败: %w", err)
	}

	// 启动进程
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("启动 FFmpeg 失败: %w", err)
	}

	// 读取原始像素数据
	reader := bufio.NewReader(output)
	pixelData := make([]byte, vr.info.Width*vr.info.Height*3)

	// 使用 io.ReadFull 确保读取完整的数据
	_, err = io.ReadFull(reader, pixelData)
	if err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("读取像素数据失败: %w", err)
	}

	// 等待进程结束
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("FFmpeg 进程异常退出: %w", err)
	}

	// 创建图像
	img := image.NewRGBA(image.Rect(0, 0, vr.info.Width, vr.info.Height))

	for y := 0; y < vr.info.Height; y++ {
		for x := 0; x < vr.info.Width; x++ {
			idx := (y*vr.info.Width + x) * 3
			r := pixelData[idx]
			g := pixelData[idx+1]
			b := pixelData[idx+2]
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	return img, nil
}

// GetInfo 获取视频信息
func (vr *VideoReader) GetInfo() *VideoInfo {
	vr.mutex.RLock()
	defer vr.mutex.RUnlock()
	return vr.info
}

// Close 关闭读取器
func (vr *VideoReader) Close() error {
	vr.mutex.Lock()
	defer vr.mutex.Unlock()

	if vr.closed {
		return nil
	}

	vr.closed = true
	vr.cancel()

	if vr.process != nil {
		vr.process.Terminate()
	}

	return nil
}

// IsClosed 检查是否已关闭
func (vr *VideoReader) IsClosed() bool {
	vr.mutex.RLock()
	defer vr.mutex.RUnlock()
	return vr.closed
}
