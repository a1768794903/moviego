package audio

import (
	"fmt"
	"time"

	"moviepy-go/pkg/core"
	"moviepy-go/pkg/ffmpeg"
)

// AudioFileClip 音频文件剪辑
type AudioFileClip struct {
	*core.BaseAudioClip
	filename   string
	reader     *ffmpeg.AudioReader
	processMgr *ffmpeg.ProcessManager
	closed     bool
}

// NewAudioFileClip 创建新的音频文件剪辑
func NewAudioFileClip(filename string, processMgr *ffmpeg.ProcessManager) *AudioFileClip {
	return &AudioFileClip{
		BaseAudioClip: core.NewBaseAudioClip(0, 0, 0, 0, 0, 0),
		filename:      filename,
		processMgr:    processMgr,
	}
}

// Open 打开音频文件
func (afc *AudioFileClip) Open() error {
	if afc.closed {
		return fmt.Errorf("剪辑已关闭")
	}

	// 创建读取器
	afc.reader = ffmpeg.NewAudioReader(afc.filename, afc.processMgr)

	// 打开音频
	if err := afc.reader.Open(); err != nil {
		return fmt.Errorf("打开音频失败: %w", err)
	}

	// 获取音频信息
	info := afc.reader.GetInfo()
	if info == nil {
		return fmt.Errorf("无法获取音频信息")
	}

	// 更新剪辑属性
	duration := time.Duration(info.Duration * float64(time.Second))
	afc.BaseAudioClip = core.NewBaseAudioClip(0, duration, duration, float64(info.SampleRate), info.Channels, info.SampleRate)

	return nil
}

// GetAudioFrame 获取音频帧
func (afc *AudioFileClip) GetAudioFrame(t time.Duration) ([]float64, error) {
	if afc.closed {
		return nil, fmt.Errorf("剪辑已关闭")
	}

	if afc.reader == nil {
		return nil, fmt.Errorf("音频未打开")
	}

	return afc.reader.GetAudioFrame(t)
}

// Subclip 创建子剪辑
func (afc *AudioFileClip) Subclip(start, end time.Duration) (core.Clip, error) {
	if afc.closed {
		return nil, fmt.Errorf("剪辑已关闭")
	}

	if start < 0 || end > afc.Duration() || start >= end {
		return nil, core.ErrInvalidTimeRange
	}

	// 创建新的子剪辑
	subclip := &AudioFileClip{
		BaseAudioClip: core.NewBaseAudioClip(start, end, end-start, afc.FPS(), afc.Channels(), afc.SampleRate()),
		filename:      afc.filename,
		processMgr:    afc.processMgr,
	}

	return subclip, nil
}

// WithSpeed 调整播放速度
func (afc *AudioFileClip) WithSpeed(factor float64) (core.Clip, error) {
	if afc.closed {
		return nil, fmt.Errorf("剪辑已关闭")
	}

	if factor <= 0 {
		return nil, core.ErrInvalidSpeedFactor
	}

	// 创建新的剪辑
	speedClip := &AudioFileClip{
		BaseAudioClip: core.NewBaseAudioClip(afc.Start(), afc.End(), afc.Duration()/time.Duration(factor*float64(time.Second)), afc.FPS()*factor, afc.Channels(), afc.SampleRate()),
		filename:      afc.filename,
		processMgr:    afc.processMgr,
	}

	return speedClip, nil
}

// WithVolume 调整音量
func (afc *AudioFileClip) WithVolume(factor float64) (core.Clip, error) {
	if afc.closed {
		return nil, fmt.Errorf("剪辑已关闭")
	}

	if factor < 0 {
		return nil, core.ErrInvalidVolumeFactor
	}

	// 创建新的剪辑
	volumeClip := &AudioFileClip{
		BaseAudioClip: core.NewBaseAudioClip(afc.Start(), afc.End(), afc.Duration(), afc.FPS(), afc.Channels(), afc.SampleRate()),
		filename:      afc.filename,
		processMgr:    afc.processMgr,
	}

	// 这里应该实现音量调整逻辑
	// 简化实现，直接返回
	return volumeClip, nil
}

// WithChannels 设置声道数
func (afc *AudioFileClip) WithChannels(channels int) (core.AudioClip, error) {
	if afc.closed {
		return nil, fmt.Errorf("剪辑已关闭")
	}

	if channels <= 0 {
		return nil, core.ErrInvalidFormat
	}

	// 创建新的剪辑
	channelsClip := &AudioFileClip{
		BaseAudioClip: core.NewBaseAudioClip(afc.Start(), afc.End(), afc.Duration(), afc.FPS(), channels, afc.SampleRate()),
		filename:      afc.filename,
		processMgr:    afc.processMgr,
	}

	return channelsClip, nil
}

// WithSampleRate 设置采样率
func (afc *AudioFileClip) WithSampleRate(sampleRate int) (core.AudioClip, error) {
	if afc.closed {
		return nil, fmt.Errorf("剪辑已关闭")
	}

	if sampleRate <= 0 {
		return nil, core.ErrInvalidFormat
	}

	// 创建新的剪辑
	sampleRateClip := &AudioFileClip{
		BaseAudioClip: core.NewBaseAudioClip(afc.Start(), afc.End(), afc.Duration(), afc.FPS(), afc.Channels(), sampleRate),
		filename:      afc.filename,
		processMgr:    afc.processMgr,
	}

	return sampleRateClip, nil
}

// Concatenate 连接音频剪辑
func (afc *AudioFileClip) Concatenate(other core.AudioClip) (core.AudioClip, error) {
	if afc.closed {
		return nil, fmt.Errorf("剪辑已关闭")
	}

	// 这里应该实现音频连接逻辑
	// 简化实现，返回错误
	return nil, core.ErrNotImplemented
}

// Mix 混合音频剪辑
func (afc *AudioFileClip) Mix(other core.AudioClip) (core.AudioClip, error) {
	if afc.closed {
		return nil, fmt.Errorf("剪辑已关闭")
	}

	// 这里应该实现音频混合逻辑
	// 简化实现，返回错误
	return nil, core.ErrNotImplemented
}

// WriteToFile 写入音频文件
func (afc *AudioFileClip) WriteToFile(filename string, options *core.WriteOptions) error {
	if afc.closed {
		return fmt.Errorf("剪辑已关闭")
	}

	// 设置默认选项
	if options == nil {
		options = &core.WriteOptions{}
	}
	if options.AudioCodec == "" {
		options.AudioCodec = "aac"
	}
	if options.AudioBitrate == "" {
		options.AudioBitrate = "128k"
	}

	// 创建音频写入器
	writerOptions := &ffmpeg.AudioWriterOptions{
		Codec:      options.AudioCodec,
		Bitrate:    options.AudioBitrate,
		SampleRate: afc.SampleRate(),
		Channels:   afc.Channels(),
	}

	writer := ffmpeg.NewAudioWriter(filename, writerOptions, afc.processMgr)

	// 打开写入器
	if err := writer.Open(); err != nil {
		return fmt.Errorf("打开写入器失败: %w", err)
	}
	defer writer.Close()

	// 计算总帧数
	totalFrames := int(afc.Duration().Seconds() * afc.FPS())
	frameInterval := time.Duration(float64(time.Second) / afc.FPS())

	fmt.Printf("开始写入音频: %s\n", filename)
	fmt.Printf("总帧数: %d, 帧间隔: %v\n", totalFrames, frameInterval)

	// 逐帧写入
	for i := 0; i < totalFrames; i++ {
		t := time.Duration(i) * frameInterval
		if t > afc.Duration() {
			break
		}

		frame, err := afc.GetAudioFrame(t)
		if err != nil {
			return fmt.Errorf("获取第 %d 帧失败: %w", i, err)
		}

		if err := writer.WriteAudioFrame(frame); err != nil {
			return fmt.Errorf("写入第 %d 帧失败: %w", i, err)
		}

		// 显示进度
		if i%100 == 0 {
			progress := float64(i) / float64(totalFrames) * 100
			fmt.Printf("进度: %.1f%% (%d/%d)\n", progress, i, totalFrames)
		}
	}

	fmt.Printf("音频写入完成: %s\n", filename)
	return nil
}

// Close 关闭剪辑
func (afc *AudioFileClip) Close() error {
	if afc.closed {
		return nil
	}

	afc.closed = true

	// 关闭读取器
	if afc.reader != nil {
		afc.reader.Close()
		afc.reader = nil
	}

	return nil
}

// IsClosed 检查是否已关闭
func (afc *AudioFileClip) IsClosed() bool {
	return afc.closed
}

// AudioInfo 音频信息
type AudioInfo struct {
	Duration   float64 `json:"duration"`
	SampleRate int     `json:"sample_rate"`
	Channels   int     `json:"channels"`
	Codec      string  `json:"codec"`
	BitRate    string  `json:"bit_rate"`
}
