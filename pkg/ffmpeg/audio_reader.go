package ffmpeg

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

// AudioInfo 音频信息
type AudioInfo struct {
	Duration   float64 `json:"duration"`
	SampleRate int     `json:"sample_rate"`
	Channels   int     `json:"channels"`
	Codec      string  `json:"codec_name"`
	BitRate    string  `json:"bit_rate"`
	Format     string  `json:"format_name"`
}

// AudioReader FFmpeg 音频读取器
type AudioReader struct {
	filename   string
	info       *AudioInfo
	processMgr *ProcessManager
	process    *ManagedProcess
	ctx        context.Context
	cancel     context.CancelFunc
	closed     bool
	mutex      sync.RWMutex
}

// NewAudioReader 创建新的音频读取器
func NewAudioReader(filename string, processMgr *ProcessManager) *AudioReader {
	ctx, cancel := context.WithCancel(context.Background())
	return &AudioReader{
		filename:   filename,
		processMgr: processMgr,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Open 打开音频文件并获取信息
func (ar *AudioReader) Open() error {
	ar.mutex.Lock()
	defer ar.mutex.Unlock()

	if ar.closed {
		return fmt.Errorf("读取器已关闭")
	}

	// 检查文件是否存在
	if _, err := os.Stat(ar.filename); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", ar.filename)
	}

	// 获取音频信息
	info, err := ar.getAudioInfo()
	if err != nil {
		return fmt.Errorf("获取音频信息失败: %w", err)
	}

	ar.info = info
	return nil
}

// getAudioInfo 获取音频信息
func (ar *AudioReader) getAudioInfo() (*AudioInfo, error) {
	args := []string{
		"-i", ar.filename,
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
			CodecName  string `json:"codec_name"`
			CodecType  string `json:"codec_type"`
			SampleRate string `json:"sample_rate"`
			Channels   int    `json:"channels"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	// 查找音频流
	var audioStream *struct {
		CodecName  string `json:"codec_name"`
		CodecType  string `json:"codec_type"`
		SampleRate string `json:"sample_rate"`
		Channels   int    `json:"channels"`
	}

	for i := range result.Streams {
		if result.Streams[i].CodecType == "audio" {
			audioStream = &result.Streams[i]
			break
		}
	}

	if audioStream == nil {
		return nil, fmt.Errorf("未找到音频流")
	}

	// 解析时长
	duration, err := strconv.ParseFloat(result.Format.Duration, 64)
	if err != nil {
		duration = 0
	}

	// 解析采样率
	sampleRate, err := strconv.Atoi(audioStream.SampleRate)
	if err != nil {
		sampleRate = 44100 // 默认采样率
	}

	return &AudioInfo{
		Duration:   duration,
		SampleRate: sampleRate,
		Channels:   audioStream.Channels,
		Codec:      audioStream.CodecName,
		BitRate:    result.Format.BitRate,
		Format:     "unknown",
	}, nil
}

// GetAudioFrame 获取指定时间的音频帧
func (ar *AudioReader) GetAudioFrame(t time.Duration) ([]float64, error) {
	ar.mutex.RLock()
	defer ar.mutex.RUnlock()

	if ar.closed {
		return nil, fmt.Errorf("读取器已关闭")
	}

	if ar.info == nil {
		return nil, fmt.Errorf("音频未打开")
	}

	// 计算时间戳
	timestamp := t.Seconds()
	if timestamp > ar.info.Duration {
		return nil, fmt.Errorf("时间超出音频长度")
	}

	// 启动 FFmpeg 进程读取音频
	args := []string{
		"-ss", fmt.Sprintf("%.3f", timestamp),
		"-i", ar.filename,
		"-t", "0.1", // 读取 0.1 秒的音频
		"-f", "f32le", // 32位浮点格式
		"-ac", strconv.Itoa(ar.info.Channels),
		"-ar", strconv.Itoa(ar.info.SampleRate),
		"-",
	}

	// 创建命令
	cmd := exec.CommandContext(ar.ctx, "ffmpeg", args...)

	// 在启动进程之前设置输出管道
	output, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("设置输出管道失败: %w", err)
	}

	// 启动进程
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("启动 FFmpeg 失败: %w", err)
	}

	// 读取音频数据
	reader := bufio.NewReader(output)
	frameSize := int(0.1 * float64(ar.info.SampleRate) * float64(ar.info.Channels))
	audioData := make([]byte, frameSize*4) // 32位浮点 = 4字节

	// 使用 io.ReadFull 确保读取完整的数据
	_, err = io.ReadFull(reader, audioData)
	if err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("读取音频数据失败: %w", err)
	}

	// 等待进程结束
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("FFmpeg 进程异常退出: %w", err)
	}

	// 转换为浮点数数组
	samples := make([]float64, frameSize)
	for i := 0; i < frameSize; i++ {
		offset := i * 4
		if offset+3 < len(audioData) {
			// 将字节转换为32位浮点数
			bits := uint32(audioData[offset]) |
				uint32(audioData[offset+1])<<8 |
				uint32(audioData[offset+2])<<16 |
				uint32(audioData[offset+3])<<24
			samples[i] = float64(int32(bits)) / float64(1<<31)
		}
	}

	return samples, nil
}

// GetInfo 获取音频信息
func (ar *AudioReader) GetInfo() *AudioInfo {
	ar.mutex.RLock()
	defer ar.mutex.RUnlock()
	return ar.info
}

// Close 关闭读取器
func (ar *AudioReader) Close() error {
	ar.mutex.Lock()
	defer ar.mutex.Unlock()

	if ar.closed {
		return nil
	}

	ar.closed = true

	// 取消上下文
	if ar.cancel != nil {
		ar.cancel()
	}

	return nil
}

// IsClosed 检查是否已关闭
func (ar *AudioReader) IsClosed() bool {
	ar.mutex.RLock()
	defer ar.mutex.RUnlock()
	return ar.closed
}
