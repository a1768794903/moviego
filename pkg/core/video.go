package core

import (
	"image"
	"time"
)

// VideoClip 视频剪辑接口
type VideoClip interface {
	Clip

	// 视频特有属性
	Size() (width, height int)
	Width() int
	Height() int
	AspectRatio() float64

	// 视频特有操作
	Resize(width, height int) (VideoClip, error)
	Rotate(angle float64) (VideoClip, error)
	Crop(x, y, width, height int) (VideoClip, error)
	WithMask(mask VideoClip) (VideoClip, error)
	WithoutMask() (VideoClip, error)
	Composite(other VideoClip, position Position) (VideoClip, error)
}

// Position 表示视频在合成中的位置
type Position struct {
	X, Y     float64
	Relative bool
}

// BaseVideoClip 视频剪辑基础实现
type BaseVideoClip struct {
	*BaseClip
	width  int
	height int
	mask   VideoClip
}

// NewBaseVideoClip 创建新的视频剪辑
func NewBaseVideoClip(start, end, duration time.Duration, fps float64, width, height int) *BaseVideoClip {
	return &BaseVideoClip{
		BaseClip: NewBaseClip(start, end, duration, fps),
		width:    width,
		height:   height,
	}
}

// Size 返回视频尺寸
func (vc *BaseVideoClip) Size() (width, height int) {
	return vc.width, vc.height
}

// Width 返回视频宽度
func (vc *BaseVideoClip) Width() int {
	return vc.width
}

// Height 返回视频高度
func (vc *BaseVideoClip) Height() int {
	return vc.height
}

// AspectRatio 返回宽高比
func (vc *BaseVideoClip) AspectRatio() float64 {
	if vc.height == 0 {
		return 0
	}
	return float64(vc.width) / float64(vc.height)
}

// Resize 调整视频尺寸
func (vc *BaseVideoClip) Resize(width, height int) (VideoClip, error) {
	if width <= 0 || height <= 0 {
		return nil, ErrInvalidFormat
	}

	// 这里应该返回一个新的视频剪辑，但基础实现返回错误
	return nil, ErrNotImplemented
}

// Rotate 旋转视频
func (vc *BaseVideoClip) Rotate(angle float64) (VideoClip, error) {
	// 这里应该返回一个新的视频剪辑，但基础实现返回错误
	return nil, ErrNotImplemented
}

// Crop 裁剪视频
func (vc *BaseVideoClip) Crop(x, y, width, height int) (VideoClip, error) {
	if x < 0 || y < 0 || width <= 0 || height <= 0 {
		return nil, ErrInvalidFormat
	}

	// 这里应该返回一个新的视频剪辑，但基础实现返回错误
	return nil, ErrNotImplemented
}

// WithMask 添加遮罩
func (vc *BaseVideoClip) WithMask(mask VideoClip) (VideoClip, error) {
	// 这里应该返回一个新的视频剪辑，但基础实现返回错误
	return nil, ErrNotImplemented
}

// WithoutMask 移除遮罩
func (vc *BaseVideoClip) WithoutMask() (VideoClip, error) {
	// 这里应该返回一个新的视频剪辑，但基础实现返回错误
	return nil, ErrNotImplemented
}

// Composite 合成视频
func (vc *BaseVideoClip) Composite(other VideoClip, position Position) (VideoClip, error) {
	// 这里应该返回一个新的视频剪辑，但基础实现返回错误
	return nil, ErrNotImplemented
}

// GetFrame 获取视频帧（基础实现）
func (vc *BaseVideoClip) GetFrame(t time.Duration) (image.Image, error) {
	// 基础实现返回黑色帧
	img := image.NewRGBA(image.Rect(0, 0, vc.width, vc.height))
	return img, nil
}
