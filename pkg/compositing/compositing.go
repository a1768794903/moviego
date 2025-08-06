package compositing

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"moviepy-go/pkg/core"
	"moviepy-go/pkg/ffmpeg"
)

// CompositeMode 合成模式
type CompositeMode int

const (
	Overlay CompositeMode = iota
	Add
	Multiply
	Screen
	Darken
	Lighten
)

// Position 位置定义
type Position struct {
	X, Y     float64
	Relative bool
	Center   bool
	Scale    float64
	Rotation float64
	Opacity  float64
}

// NewPosition 创建新位置
func NewPosition(x, y float64) *Position {
	return &Position{
		X:        x,
		Y:        y,
		Relative: false,
		Center:   false,
		Scale:    1.0,
		Rotation: 0.0,
		Opacity:  1.0,
	}
}

// NewCenteredPosition 创建居中位置
func NewCenteredPosition() *Position {
	return &Position{
		X:        0,
		Y:        0,
		Relative: false,
		Center:   true,
		Scale:    1.0,
		Rotation: 0.0,
		Opacity:  1.0,
	}
}

// CompositeVideoClip 合成视频剪辑
type CompositeVideoClip struct {
	*core.BaseVideoClip
	clips      []core.VideoClip
	positions  []*Position
	mode       CompositeMode
	processMgr *ffmpeg.ProcessManager
	closed     bool
}

// NewCompositeVideoClip 创建新的合成视频剪辑
func NewCompositeVideoClip(clips []core.VideoClip, positions []*Position, mode CompositeMode, processMgr *ffmpeg.ProcessManager) *CompositeVideoClip {
	if len(clips) == 0 {
		return nil
	}

	baseClip := clips[0]
	width := baseClip.Width()
	height := baseClip.Height()

	maxDuration := baseClip.Duration()
	for _, clip := range clips {
		if clip.Duration() > maxDuration {
			maxDuration = clip.Duration()
		}
	}

	return &CompositeVideoClip{
		BaseVideoClip: core.NewBaseVideoClip(0, maxDuration, maxDuration, baseClip.FPS(), width, height),
		clips:         clips,
		positions:     positions,
		mode:          mode,
		processMgr:    processMgr,
	}
}

// GetFrame 获取合成帧
func (cvc *CompositeVideoClip) GetFrame(t time.Duration) (image.Image, error) {
	if cvc.closed {
		return nil, fmt.Errorf("剪辑已关闭")
	}

	if len(cvc.clips) == 0 {
		return nil, fmt.Errorf("没有可合成的剪辑")
	}

	baseFrame, err := cvc.clips[0].GetFrame(t)
	if err != nil {
		return nil, fmt.Errorf("获取基础帧失败: %w", err)
	}

	composite := image.NewRGBA(baseFrame.Bounds())
	bounds := baseFrame.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			composite.Set(x, y, baseFrame.At(x, y))
		}
	}

	for i := 1; i < len(cvc.clips); i++ {
		clip := cvc.clips[i]
		position := cvc.positions[i]

		clipFrame, err := clip.GetFrame(t)
		if err != nil {
			continue
		}

		transformedFrame, err := cvc.applyTransform(clipFrame, position)
		if err != nil {
			continue
		}

		cvc.compositeFrame(composite, transformedFrame, position, cvc.mode)
	}

	return composite, nil
}

// applyTransform 应用位置变换
func (cvc *CompositeVideoClip) applyTransform(frame image.Image, position *Position) (image.Image, error) {
	bounds := frame.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	targetWidth := int(float64(width) * position.Scale)
	targetHeight := int(float64(height) * position.Scale)

	transformed := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	for y := 0; y < targetHeight; y++ {
		for x := 0; x < targetWidth; x++ {
			srcX := int(float64(x) * float64(width) / float64(targetWidth))
			srcY := int(float64(y) * float64(height) / float64(targetHeight))

			if srcX >= width {
				srcX = width - 1
			}
			if srcY >= height {
				srcY = height - 1
			}

			transformed.Set(x, y, frame.At(srcX, srcY))
		}
	}

	return transformed, nil
}

// compositeFrame 合成帧
func (cvc *CompositeVideoClip) compositeFrame(base, overlay image.Image, position *Position, mode CompositeMode) {
	baseBounds := base.Bounds()
	overlayBounds := overlay.Bounds()

	offsetX, offsetY := cvc.calculateOffset(baseBounds, overlayBounds, position)

	for y := overlayBounds.Min.Y; y < overlayBounds.Max.Y; y++ {
		for x := overlayBounds.Min.X; x < overlayBounds.Max.X; x++ {
			targetX := offsetX + x
			targetY := offsetY + y

			if targetX < baseBounds.Min.X || targetX >= baseBounds.Max.X ||
				targetY < baseBounds.Min.Y || targetY >= baseBounds.Max.Y {
				continue
			}

			baseColor := base.At(targetX, targetY)
			overlayColor := overlay.At(x, y)

			if position.Opacity < 1.0 {
				overlayColor = cvc.applyOpacity(overlayColor, position.Opacity)
			}

			compositeColor := cvc.blendColors(baseColor, overlayColor, mode)
			base.(*image.RGBA).Set(targetX, targetY, compositeColor)
		}
	}
}

// calculateOffset 计算偏移量
func (cvc *CompositeVideoClip) calculateOffset(baseBounds, overlayBounds image.Rectangle, position *Position) (int, int) {
	baseWidth := baseBounds.Dx()
	baseHeight := baseBounds.Dy()
	overlayWidth := overlayBounds.Dx()
	overlayHeight := overlayBounds.Dy()

	var offsetX, offsetY int

	if position.Center {
		offsetX = (baseWidth - overlayWidth) / 2
		offsetY = (baseHeight - overlayHeight) / 2
	} else {
		offsetX = int(position.X)
		offsetY = int(position.Y)
	}

	return offsetX, offsetY
}

// applyOpacity 应用透明度
func (cvc *CompositeVideoClip) applyOpacity(color color.Color, opacity float64) color.Color {
	// 简化实现，直接返回原颜色
	return color
}

// blendColors 混合颜色
func (cvc *CompositeVideoClip) blendColors(base, overlay color.Color, mode CompositeMode) color.Color {
	r1, g1, b1, _ := base.RGBA()
	r2, g2, b2, a2 := overlay.RGBA()

	var r, g, b uint32

	switch mode {
	case Overlay:
		r = cvc.blendOverlay(r1, r2)
		g = cvc.blendOverlay(g1, g2)
		b = cvc.blendOverlay(b1, b2)
	case Add:
		r = cvc.clamp(r1 + r2)
		g = cvc.clamp(g1 + g2)
		b = cvc.clamp(b1 + b2)
	case Multiply:
		r = (r1 * r2) / 65535
		g = (g1 * g2) / 65535
		b = (b1 * b2) / 65535
	case Screen:
		r = cvc.blendScreen(r1, r2)
		g = cvc.blendScreen(g1, g2)
		b = cvc.blendScreen(b1, b2)
	case Darken:
		r = cvc.min(r1, r2)
		g = cvc.min(g1, g2)
		b = cvc.min(b1, b2)
	case Lighten:
		r = cvc.max(r1, r2)
		g = cvc.max(g1, g2)
		b = cvc.max(b1, b2)
	default:
		r = r2
		g = g2
		b = b2
	}

	return color.RGBA64{
		R: uint16(r),
		G: uint16(g),
		B: uint16(b),
		A: uint16(a2),
	}
}

// blendOverlay 叠加混合
func (cvc *CompositeVideoClip) blendOverlay(base, overlay uint32) uint32 {
	if base < 32768 {
		return (2 * base * overlay) / 65535
	}
	return 65535 - (2*(65535-base)*(65535-overlay))/65535
}

// blendScreen 屏幕混合
func (cvc *CompositeVideoClip) blendScreen(base, overlay uint32) uint32 {
	return 65535 - ((65535-base)*(65535-overlay))/65535
}

// clamp 限制值范围
func (cvc *CompositeVideoClip) clamp(value uint32) uint32 {
	if value > 65535 {
		return 65535
	}
	return value
}

// min 最小值
func (cvc *CompositeVideoClip) min(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

// max 最大值
func (cvc *CompositeVideoClip) max(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

// GetAudioFrame 获取音频帧
func (cvc *CompositeVideoClip) GetAudioFrame(t time.Duration) ([]float64, error) {
	if cvc.closed {
		return nil, fmt.Errorf("剪辑已关闭")
	}

	if len(cvc.clips) > 0 {
		return cvc.clips[0].GetAudioFrame(t)
	}

	return nil, fmt.Errorf("没有音频")
}

// Subclip 创建子剪辑
func (cvc *CompositeVideoClip) Subclip(start, end time.Duration) (core.Clip, error) {
	if start < 0 || end > cvc.Duration() || start >= end {
		return nil, core.ErrInvalidTimeRange
	}

	// 创建子剪辑
	subclips := make([]core.VideoClip, len(cvc.clips))
	for i, clip := range cvc.clips {
		subclip, err := clip.Subclip(start, end)
		if err != nil {
			return nil, fmt.Errorf("创建子剪辑失败: %w", err)
		}

		videoSubclip, ok := subclip.(core.VideoClip)
		if !ok {
			return nil, fmt.Errorf("子剪辑不是视频剪辑")
		}
		subclips[i] = videoSubclip
	}

	return NewCompositeVideoClip(subclips, cvc.positions, cvc.mode, cvc.processMgr), nil
}

// WithSpeed 调整播放速度
func (cvc *CompositeVideoClip) WithSpeed(factor float64) (core.Clip, error) {
	if factor <= 0 {
		return nil, core.ErrInvalidSpeedFactor
	}

	// 创建速度调整的剪辑
	speedClips := make([]core.VideoClip, len(cvc.clips))
	for i, clip := range cvc.clips {
		speedClip, err := clip.WithSpeed(factor)
		if err != nil {
			return nil, fmt.Errorf("调整剪辑速度失败: %w", err)
		}

		videoSpeedClip, ok := speedClip.(core.VideoClip)
		if !ok {
			return nil, fmt.Errorf("速度剪辑不是视频剪辑")
		}
		speedClips[i] = videoSpeedClip
	}

	return NewCompositeVideoClip(speedClips, cvc.positions, cvc.mode, cvc.processMgr), nil
}

// WithVolume 调整音量
func (cvc *CompositeVideoClip) WithVolume(factor float64) (core.Clip, error) {
	if factor < 0 {
		return nil, core.ErrInvalidVolumeFactor
	}

	// 创建音量调整的剪辑
	volumeClips := make([]core.VideoClip, len(cvc.clips))
	for i, clip := range cvc.clips {
		volumeClip, err := clip.WithVolume(factor)
		if err != nil {
			return nil, fmt.Errorf("调整剪辑音量失败: %w", err)
		}

		videoVolumeClip, ok := volumeClip.(core.VideoClip)
		if !ok {
			return nil, fmt.Errorf("音量剪辑不是视频剪辑")
		}
		volumeClips[i] = videoVolumeClip
	}

	return NewCompositeVideoClip(volumeClips, cvc.positions, cvc.mode, cvc.processMgr), nil
}

// WithAudio 添加音频
func (cvc *CompositeVideoClip) WithAudio(audio core.AudioClip) (core.Clip, error) {
	// 简化实现，直接返回原剪辑
	return cvc, nil
}

// WithoutAudio 移除音频
func (cvc *CompositeVideoClip) WithoutAudio() (core.Clip, error) {
	// 简化实现，直接返回原剪辑
	return cvc, nil
}

// WriteToFile 写入文件
func (cvc *CompositeVideoClip) WriteToFile(filename string, options *core.WriteOptions) error {
	if cvc.closed {
		return fmt.Errorf("剪辑已关闭")
	}

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
		options.FPS = cvc.FPS()
	}

	writerOptions := &ffmpeg.VideoWriterOptions{
		Codec:   options.Codec,
		Bitrate: options.Bitrate,
		FPS:     options.FPS,
	}

	writer := ffmpeg.NewVideoWriter(filename, cvc.Width(), cvc.Height(), writerOptions, cvc.processMgr)

	if err := writer.Open(); err != nil {
		return fmt.Errorf("打开写入器失败: %w", err)
	}
	defer writer.Close()

	totalFrames := int(cvc.Duration().Seconds() * options.FPS)
	frameInterval := time.Duration(float64(time.Second) / options.FPS)

	fmt.Printf("开始写入合成视频: %s\n", filename)
	fmt.Printf("剪辑数量: %d\n", len(cvc.clips))
	fmt.Printf("合成模式: %d\n", cvc.mode)
	fmt.Printf("总帧数: %d, 帧间隔: %v\n", totalFrames, frameInterval)

	for i := 0; i < totalFrames; i++ {
		t := time.Duration(i) * frameInterval
		if t > cvc.Duration() {
			break
		}

		frame, err := cvc.GetFrame(t)
		if err != nil {
			return fmt.Errorf("获取第 %d 帧失败: %w", i, err)
		}

		if err := writer.WriteFrame(frame); err != nil {
			return fmt.Errorf("写入第 %d 帧失败: %w", i, err)
		}

		if i%100 == 0 {
			progress := float64(i) / float64(totalFrames) * 100
			fmt.Printf("进度: %.1f%% (%d/%d)\n", progress, i, totalFrames)
		}
	}

	fmt.Printf("合成视频写入完成: %s\n", filename)
	return nil
}

// Close 关闭剪辑
func (cvc *CompositeVideoClip) Close() error {
	if cvc.closed {
		return nil
	}
	cvc.closed = true

	// 不关闭子剪辑，让调用者管理剪辑的生命周期
	// 这样可以避免多个合成操作之间的剪辑关闭冲突
	// for _, clip := range cvc.clips {
	// 	if closer, ok := clip.(interface{ Close() error }); ok {
	// 		closer.Close()
	// 	}
	// }

	return nil
}

// GetClips 获取所有剪辑
func (cvc *CompositeVideoClip) GetClips() []core.VideoClip {
	return cvc.clips
}

// GetPositions 获取所有位置
func (cvc *CompositeVideoClip) GetPositions() []*Position {
	return cvc.positions
}

// GetMode 获取合成模式
func (cvc *CompositeVideoClip) GetMode() CompositeMode {
	return cvc.mode
}
