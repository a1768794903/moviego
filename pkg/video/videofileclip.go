package video

import (
	"fmt"
	"image"
	"time"

	"moviepy-go/pkg/audio"
	"moviepy-go/pkg/core"
	"moviepy-go/pkg/ffmpeg"
)

// VideoFileClip 视频文件剪辑
type VideoFileClip struct {
	*core.BaseVideoClip
	filename    string
	reader      *ffmpeg.VideoReader
	processMgr  *ffmpeg.ProcessManager
	audio       core.AudioClip
	closed      bool
	speedFactor float64 // 速度调整因子，1.0表示正常速度
}

// NewVideoFileClip 创建新的视频文件剪辑
func NewVideoFileClip(filename string, processMgr *ffmpeg.ProcessManager) *VideoFileClip {
	return &VideoFileClip{
		BaseVideoClip: core.NewBaseVideoClip(0, 0, 0, 0, 0, 0),
		filename:      filename,
		processMgr:    processMgr,
		speedFactor:   1.0, // 默认正常速度
	}
}

// Open 打开视频文件
func (vfc *VideoFileClip) Open() error {
	if vfc.closed {
		return fmt.Errorf("剪辑已关闭")
	}

	// 创建读取器
	vfc.reader = ffmpeg.NewVideoReader(vfc.filename, vfc.processMgr)

	// 打开视频
	if err := vfc.reader.Open(); err != nil {
		return fmt.Errorf("打开视频失败: %w", err)
	}

	// 获取视频信息
	info := vfc.reader.GetInfo()
	if info == nil {
		return fmt.Errorf("无法获取视频信息")
	}

	// 更新剪辑属性
	duration := time.Duration(info.Duration * float64(time.Second))
	vfc.BaseVideoClip = core.NewBaseVideoClip(0, duration, duration, info.FPS, info.Width, info.Height)

	// 如果有音频，创建音频剪辑
	if info.HasAudio {
		audioClip := audio.NewAudioFileClip(vfc.filename, vfc.processMgr)
		if err := audioClip.Open(); err == nil {
			vfc.audio = audioClip
		}
	}

	return nil
}

// GetFrame 获取指定时间的帧
func (vfc *VideoFileClip) GetFrame(t time.Duration) (image.Image, error) {
	if vfc.closed {
		return nil, fmt.Errorf("剪辑已关闭")
	}

	if vfc.reader == nil {
		return nil, fmt.Errorf("视频未打开")
	}

	// 对于子剪辑，需要调整时间偏移
	absoluteTime := vfc.Start() + t

	// 对于速度调整，需要调整时间映射
	if vfc.speedFactor != 1.0 && vfc.speedFactor != 0 {
		// 速度调整：将当前时间映射到原视频的时间
		// 例如：2倍速时，t=1s应该获取原视频t=2s的帧
		absoluteTime = vfc.Start() + time.Duration(float64(t)*vfc.speedFactor)
	}

	return vfc.reader.GetFrame(absoluteTime)
}

// GetAudioFrame 获取指定时间的音频帧
func (vfc *VideoFileClip) GetAudioFrame(t time.Duration) ([]float64, error) {
	if vfc.closed {
		return nil, fmt.Errorf("剪辑已关闭")
	}

	if vfc.audio == nil {
		// 返回静音
		sampleRate := int(vfc.FPS())
		if sampleRate == 0 {
			sampleRate = 44100
		}
		return make([]float64, sampleRate), nil
	}

	return vfc.audio.GetAudioFrame(t)
}

// Subclip 创建子剪辑
func (vfc *VideoFileClip) Subclip(start, end time.Duration) (core.Clip, error) {
	if start < 0 || end > vfc.Duration() || start >= end {
		return nil, core.ErrInvalidTimeRange
	}

	// 创建新的子剪辑
	subclip := &VideoFileClip{
		BaseVideoClip: core.NewBaseVideoClip(start, end, end-start, vfc.FPS(), vfc.Width(), vfc.Height()),
		filename:      vfc.filename,
		processMgr:    vfc.processMgr,
		audio:         vfc.audio,
		reader:        vfc.reader, // 共享同一个读取器
		closed:        false,
		speedFactor:   vfc.speedFactor, // 继承速度因子
	}

	// 子剪辑不需要重新打开，因为它共享父剪辑的读取器
	return subclip, nil
}

// WithSpeed 调整播放速度
func (vfc *VideoFileClip) WithSpeed(factor float64) (core.Clip, error) {
	if factor <= 0 {
		return nil, core.ErrInvalidSpeedFactor
	}

	// 计算新的持续时间：速度加快时间变短，速度减慢时间变长
	newDuration := time.Duration(float64(vfc.Duration()) / factor)

	// 创建新的剪辑
	speedClip := &VideoFileClip{
		BaseVideoClip: core.NewBaseVideoClip(vfc.Start(), vfc.Start()+newDuration, newDuration, vfc.FPS(), vfc.Width(), vfc.Height()),
		filename:      vfc.filename,
		processMgr:    vfc.processMgr,
		audio:         vfc.audio,
		reader:        vfc.reader, // 共享同一个读取器
		closed:        false,
		speedFactor:   factor, // 添加速度因子字段
	}

	return speedClip, nil
}

// WithVolume 调整音量
func (vfc *VideoFileClip) WithVolume(factor float64) (core.Clip, error) {
	if factor < 0 {
		return nil, core.ErrInvalidVolumeFactor
	}

	// 创建新的剪辑
	volumeClip := &VideoFileClip{
		BaseVideoClip: core.NewBaseVideoClip(vfc.Start(), vfc.End(), vfc.Duration(), vfc.FPS(), vfc.Width(), vfc.Height()),
		filename:      vfc.filename,
		processMgr:    vfc.processMgr,
		audio:         vfc.audio,  // 这里应该创建音量调整后的音频
		reader:        vfc.reader, // 共享同一个读取器
		closed:        false,
		speedFactor:   vfc.speedFactor, // 继承速度因子
	}

	return volumeClip, nil
}

// WithAudio 添加音频
func (vfc *VideoFileClip) WithAudio(audio core.AudioClip) (core.Clip, error) {
	// 创建新的剪辑
	audioClip := &VideoFileClip{
		BaseVideoClip: core.NewBaseVideoClip(vfc.Start(), vfc.End(), vfc.Duration(), vfc.FPS(), vfc.Width(), vfc.Height()),
		filename:      vfc.filename,
		processMgr:    vfc.processMgr,
		audio:         audio,
		reader:        vfc.reader, // 共享同一个读取器
		closed:        false,
		speedFactor:   vfc.speedFactor, // 继承速度因子
	}

	return audioClip, nil
}

// WithoutAudio 移除音频
func (vfc *VideoFileClip) WithoutAudio() (core.Clip, error) {
	// 创建新的剪辑
	noAudioClip := &VideoFileClip{
		BaseVideoClip: core.NewBaseVideoClip(vfc.Start(), vfc.End(), vfc.Duration(), vfc.FPS(), vfc.Width(), vfc.Height()),
		filename:      vfc.filename,
		processMgr:    vfc.processMgr,
		audio:         nil,
		reader:        vfc.reader, // 共享同一个读取器
		closed:        false,
		speedFactor:   vfc.speedFactor, // 继承速度因子
	}

	return noAudioClip, nil
}

// WriteToFile 写入文件
func (vfc *VideoFileClip) WriteToFile(filename string, options *core.WriteOptions) error {
	if vfc.closed {
		return fmt.Errorf("剪辑已关闭")
	}

	// 设置默认选项
	if options == nil {
		options = &core.WriteOptions{}
	}
	if options.Codec == "" {
		options.Codec = "libx264"
	}
	if options.Bitrate == "" {
		options.Bitrate = "1000k" // 降低比特率以提高兼容性
	}
	if options.FPS == 0 {
		options.FPS = vfc.FPS()
	}

	// 创建视频写入器
	writerOptions := &ffmpeg.VideoWriterOptions{
		Codec:   options.Codec,
		Bitrate: options.Bitrate,
		FPS:     options.FPS,
	}

	writer := ffmpeg.NewVideoWriter(filename, vfc.Width(), vfc.Height(), writerOptions, vfc.processMgr)

	// 打开写入器
	if err := writer.Open(); err != nil {
		return fmt.Errorf("打开写入器失败: %w", err)
	}
	defer writer.Close()

	// 计算总帧数
	totalFrames := int(vfc.Duration().Seconds() * options.FPS)
	frameInterval := time.Duration(float64(time.Second) / options.FPS)

	fmt.Printf("开始写入视频: %s\n", filename)
	fmt.Printf("总帧数: %d, 帧间隔: %v\n", totalFrames, frameInterval)

	// 逐帧写入
	for i := 0; i < totalFrames; i++ {
		t := time.Duration(i) * frameInterval
		if t > vfc.Duration() {
			break
		}

		frame, err := vfc.GetFrame(t)
		if err != nil {
			return fmt.Errorf("获取第 %d 帧失败: %w", i, err)
		}

		if err := writer.WriteFrame(frame); err != nil {
			return fmt.Errorf("写入第 %d 帧失败: %w", i, err)
		}

		// 显示进度
		if i%100 == 0 {
			progress := float64(i) / float64(totalFrames) * 100
			fmt.Printf("进度: %.1f%% (%d/%d)\n", progress, i, totalFrames)
		}
	}

	fmt.Printf("视频写入完成: %s\n", filename)
	return nil
}

// Close 关闭剪辑
func (vfc *VideoFileClip) Close() error {
	if vfc.closed {
		return nil
	}

	vfc.closed = true

	// 关闭读取器
	if vfc.reader != nil {
		// 读取器没有 Close 方法，但我们可以标记为关闭
		vfc.reader = nil
	}

	// 关闭音频
	if vfc.audio != nil {
		vfc.audio.Close()
		vfc.audio = nil
	}

	return nil
}
