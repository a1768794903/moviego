package video

import (
	"fmt"
	"image"
	"time"

	"moviepy-go/pkg/core"
	"moviepy-go/pkg/effects"
	"moviepy-go/pkg/ffmpeg"
)

// EffectVideoClip 支持特效的视频剪辑
type EffectVideoClip struct {
	*core.BaseVideoClip
	originalClip core.VideoClip
	effects      []effects.VideoEffect
	processMgr   *ffmpeg.ProcessManager
	closed       bool
}

// NewEffectVideoClip 创建新的特效视频剪辑
func NewEffectVideoClip(original core.VideoClip, processMgr *ffmpeg.ProcessManager) *EffectVideoClip {
	return &EffectVideoClip{
		BaseVideoClip: core.NewBaseVideoClip(original.Start(), original.End(), original.Duration(), original.FPS(), original.Width(), original.Height()),
		originalClip:  original,
		effects:       make([]effects.VideoEffect, 0),
		processMgr:    processMgr,
	}
}

// AddEffect 添加特效
func (evc *EffectVideoClip) AddEffect(effect effects.VideoEffect) {
	evc.effects = append(evc.effects, effect)

	// 重新计算应用所有特效后的最终尺寸
	evc.updateFinalDimensions()
}

// updateFinalDimensions 更新应用所有特效后的最终尺寸
func (evc *EffectVideoClip) updateFinalDimensions() {
	// 从原始剪辑尺寸开始
	width := evc.originalClip.Width()
	height := evc.originalClip.Height()

	// 依次应用每个特效来计算最终尺寸
	for _, effect := range evc.effects {
		width, height = evc.calculateEffectDimensions(effect, width, height)
	}

	// 如果尺寸有变化，更新BaseVideoClip
	if width != evc.Width() || height != evc.Height() {
		evc.BaseVideoClip = core.NewBaseVideoClip(
			evc.Start(), evc.End(), evc.Duration(), evc.FPS(),
			width, height,
		)
	}
}

// calculateEffectDimensions 计算特效应用到指定尺寸后的新尺寸
func (evc *EffectVideoClip) calculateEffectDimensions(effect effects.VideoEffect, inputWidth, inputHeight int) (int, int) {
	// 创建测试图像
	testImg := image.NewRGBA(image.Rect(0, 0, inputWidth, inputHeight))

	// 应用特效
	resultImg, err := effect.ApplyToFrame(testImg)
	if err != nil {
		// 如果出错，返回输入尺寸
		return inputWidth, inputHeight
	}

	bounds := resultImg.Bounds()
	return bounds.Dx(), bounds.Dy()
}

// GetFrame 获取帧，应用所有特效
func (evc *EffectVideoClip) GetFrame(t time.Duration) (image.Image, error) {
	if evc.closed {
		return nil, fmt.Errorf("剪辑已关闭")
	}

	// 从原始剪辑获取帧
	frame, err := evc.originalClip.GetFrame(t)
	if err != nil {
		return nil, fmt.Errorf("获取原始帧失败: %w", err)
	}

	// 应用所有特效
	result := frame
	for _, effect := range evc.effects {
		result, err = effect.ApplyToFrame(result)
		if err != nil {
			return nil, fmt.Errorf("应用特效 %s 失败: %w", effect.GetName(), err)
		}
	}

	return result, nil
}

// GetAudioFrame 获取音频帧
func (evc *EffectVideoClip) GetAudioFrame(t time.Duration) ([]float64, error) {
	if evc.closed {
		return nil, fmt.Errorf("剪辑已关闭")
	}

	// 如果有音频，从原始剪辑获取
	if audioClip, ok := evc.originalClip.(core.Clip); ok {
		return audioClip.GetAudioFrame(t)
	}

	return nil, fmt.Errorf("原始剪辑不支持音频")
}

// Subclip 创建子剪辑
func (evc *EffectVideoClip) Subclip(start, end time.Duration) (core.Clip, error) {
	if start < 0 || end > evc.Duration() || start >= end {
		return nil, core.ErrInvalidTimeRange
	}

	// 创建原始剪辑的子剪辑
	originalSubclip, err := evc.originalClip.Subclip(start, end)
	if err != nil {
		return nil, fmt.Errorf("创建原始子剪辑失败: %w", err)
	}

	// 转换为视频剪辑
	videoSubclip, ok := originalSubclip.(core.VideoClip)
	if !ok {
		return nil, fmt.Errorf("原始子剪辑不是视频剪辑")
	}

	// 创建新的特效剪辑
	effectSubclip := NewEffectVideoClip(videoSubclip, evc.processMgr)

	// 复制特效
	for _, effect := range evc.effects {
		effectSubclip.AddEffect(effect)
	}

	return effectSubclip, nil
}

// WithSpeed 调整播放速度
func (evc *EffectVideoClip) WithSpeed(factor float64) (core.Clip, error) {
	if factor <= 0 {
		return nil, core.ErrInvalidSpeedFactor
	}

	// 创建原始剪辑的速度调整版本
	originalSpeedClip, err := evc.originalClip.WithSpeed(factor)
	if err != nil {
		return nil, fmt.Errorf("调整原始剪辑速度失败: %w", err)
	}

	// 转换为视频剪辑
	videoSpeedClip, ok := originalSpeedClip.(core.VideoClip)
	if !ok {
		return nil, fmt.Errorf("原始速度剪辑不是视频剪辑")
	}

	// 创建新的特效剪辑
	effectSpeedClip := NewEffectVideoClip(videoSpeedClip, evc.processMgr)

	// 复制特效
	for _, effect := range evc.effects {
		effectSpeedClip.AddEffect(effect)
	}

	return effectSpeedClip, nil
}

// WithVolume 调整音量
func (evc *EffectVideoClip) WithVolume(factor float64) (core.Clip, error) {
	if factor < 0 {
		return nil, core.ErrInvalidVolumeFactor
	}

	// 创建原始剪辑的音量调整版本
	originalVolumeClip, err := evc.originalClip.WithVolume(factor)
	if err != nil {
		return nil, fmt.Errorf("调整原始剪辑音量失败: %w", err)
	}

	// 转换为视频剪辑
	videoVolumeClip, ok := originalVolumeClip.(core.VideoClip)
	if !ok {
		return nil, fmt.Errorf("原始音量剪辑不是视频剪辑")
	}

	// 创建新的特效剪辑
	effectVolumeClip := NewEffectVideoClip(videoVolumeClip, evc.processMgr)

	// 复制特效
	for _, effect := range evc.effects {
		effectVolumeClip.AddEffect(effect)
	}

	return effectVolumeClip, nil
}

// WithAudio 添加音频
func (evc *EffectVideoClip) WithAudio(audio core.AudioClip) (core.Clip, error) {
	// 创建原始剪辑的音频版本
	originalAudioClip, err := evc.originalClip.WithAudio(audio)
	if err != nil {
		return nil, fmt.Errorf("为原始剪辑添加音频失败: %w", err)
	}

	// 转换为视频剪辑
	videoAudioClip, ok := originalAudioClip.(core.VideoClip)
	if !ok {
		return nil, fmt.Errorf("原始音频剪辑不是视频剪辑")
	}

	// 创建新的特效剪辑
	effectAudioClip := NewEffectVideoClip(videoAudioClip, evc.processMgr)

	// 复制特效
	for _, effect := range evc.effects {
		effectAudioClip.AddEffect(effect)
	}

	return effectAudioClip, nil
}

// WithoutAudio 移除音频
func (evc *EffectVideoClip) WithoutAudio() (core.Clip, error) {
	// 创建原始剪辑的无音频版本
	originalNoAudioClip, err := evc.originalClip.WithoutAudio()
	if err != nil {
		return nil, fmt.Errorf("移除原始剪辑音频失败: %w", err)
	}

	// 转换为视频剪辑
	videoNoAudioClip, ok := originalNoAudioClip.(core.VideoClip)
	if !ok {
		return nil, fmt.Errorf("原始无音频剪辑不是视频剪辑")
	}

	// 创建新的特效剪辑
	effectNoAudioClip := NewEffectVideoClip(videoNoAudioClip, evc.processMgr)

	// 复制特效
	for _, effect := range evc.effects {
		effectNoAudioClip.AddEffect(effect)
	}

	return effectNoAudioClip, nil
}

// WriteToFile 写入文件
func (evc *EffectVideoClip) WriteToFile(filename string, options *core.WriteOptions) error {
	if evc.closed {
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
		options.Bitrate = "2000k"
	}
	if options.FPS == 0 {
		options.FPS = evc.FPS()
	}

	// 创建视频写入器
	writerOptions := &ffmpeg.VideoWriterOptions{
		Codec:   options.Codec,
		Bitrate: options.Bitrate,
		FPS:     options.FPS,
	}

	writer := ffmpeg.NewVideoWriter(filename, evc.Width(), evc.Height(), writerOptions, evc.processMgr)

	// 打开写入器
	if err := writer.Open(); err != nil {
		return fmt.Errorf("打开写入器失败: %w", err)
	}
	defer writer.Close()

	// 计算总帧数
	totalFrames := int(evc.Duration().Seconds() * options.FPS)
	frameInterval := time.Duration(float64(time.Second) / options.FPS)

	fmt.Printf("开始写入特效视频: %s\n", filename)
	fmt.Printf("特效数量: %d\n", len(evc.effects))
	for i, effect := range evc.effects {
		fmt.Printf("  特效 %d: %s\n", i+1, effect.GetName())
	}
	fmt.Printf("总帧数: %d, 帧间隔: %v\n", totalFrames, frameInterval)

	// 逐帧写入
	for i := 0; i < totalFrames; i++ {
		t := time.Duration(i) * frameInterval
		if t > evc.Duration() {
			break
		}

		frame, err := evc.GetFrame(t)
		if err != nil {
			return fmt.Errorf("获取第 %d 帧失败: %w", i, err)
		}

		// 检查帧尺寸
		bounds := frame.Bounds()
		if bounds.Dx() != evc.Width() || bounds.Dy() != evc.Height() {
			fmt.Printf("警告: 第 %d 帧尺寸不匹配，期望 %dx%d，实际 %dx%d\n",
				i, evc.Width(), evc.Height(), bounds.Dx(), bounds.Dy())
		}

		if err := writer.WriteFrame(frame); err != nil {
			return fmt.Errorf("写入第 %d 帧失败: %w", i, err)
		}

		// 显示进度
		if i%10 == 0 || i < 10 { // 前10帧每帧显示，之后每10帧显示
			progress := float64(i) / float64(totalFrames) * 100
			fmt.Printf("进度: %.1f%% (%d/%d)\n", progress, i, totalFrames)
		}
	}

	fmt.Printf("特效视频写入完成: %s\n", filename)
	return nil
}

// Close 关闭剪辑
func (evc *EffectVideoClip) Close() error {
	if evc.closed {
		return nil
	}
	evc.closed = true

	// 不关闭原始剪辑，让调用者管理剪辑的生命周期
	// 这样可以避免多个特效操作之间的剪辑关闭冲突
	// if evc.originalClip != nil {
	// 	if closer, ok := evc.originalClip.(interface{ Close() error }); ok {
	// 		closer.Close()
	// 	}
	// 	evc.originalClip = nil
	// }

	return nil
}

// GetEffects 获取所有特效
func (evc *EffectVideoClip) GetEffects() []effects.VideoEffect {
	return evc.effects
}

// ClearEffects 清除所有特效
func (evc *EffectVideoClip) ClearEffects() {
	evc.effects = make([]effects.VideoEffect, 0)
}
