package effects

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"moviepy-go/pkg/core"
)

// Effect 特效接口
type Effect interface {
	// Apply 应用特效到剪辑
	Apply(clip core.Clip) (core.Clip, error)

	// GetName 获取特效名称
	GetName() string
}

// VideoEffect 视频特效接口
type VideoEffect interface {
	Effect

	// ApplyToFrame 应用特效到单个帧
	ApplyToFrame(frame image.Image) (image.Image, error)
}

// AudioEffect 音频特效接口
type AudioEffect interface {
	Effect

	// ApplyToAudioFrame 应用特效到音频帧
	ApplyToAudioFrame(samples []float64) ([]float64, error)
}

// TransformEffect 变换特效基础结构
type TransformEffect struct {
	name string
}

// GetName 获取特效名称
func (te *TransformEffect) GetName() string {
	return te.name
}

// ResizeEffect 缩放特效
type ResizeEffect struct {
	TransformEffect
	width  int
	height int
}

// NewResizeEffect 创建缩放特效，自动调整为偶数尺寸
func NewResizeEffect(width, height int) *ResizeEffect {
	// 确保尺寸是偶数（H.264编码器要求）
	if width%2 != 0 {
		width++
	}
	if height%2 != 0 {
		height++
	}

	return &ResizeEffect{
		TransformEffect: TransformEffect{name: "resize"},
		width:           width,
		height:          height,
	}
}

// Apply 应用缩放特效
func (re *ResizeEffect) Apply(clip core.Clip) (core.Clip, error) {
	// 这里应该返回一个新的剪辑，应用了缩放特效
	// 简化实现，直接返回原剪辑
	return clip, nil
}

// ApplyToFrame 应用缩放特效到帧
func (re *ResizeEffect) ApplyToFrame(frame image.Image) (image.Image, error) {
	bounds := frame.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	// 创建目标图像
	dst := image.NewRGBA(image.Rect(0, 0, re.width, re.height))

	// 简单的最近邻缩放算法
	for y := 0; y < re.height; y++ {
		for x := 0; x < re.width; x++ {
			// 计算源坐标
			srcX := int(float64(x) * float64(srcWidth) / float64(re.width))
			srcY := int(float64(y) * float64(srcHeight) / float64(re.height))

			// 确保坐标在边界内
			if srcX >= srcWidth {
				srcX = srcWidth - 1
			}
			if srcY >= srcHeight {
				srcY = srcHeight - 1
			}

			// 复制像素
			dst.Set(x, y, frame.At(srcX, srcY))
		}
	}

	return dst, nil
}

// RotateEffect 旋转特效
type RotateEffect struct {
	TransformEffect
	angle float64 // 角度，以度为单位
}

// NewRotateEffect 创建旋转特效
func NewRotateEffect(angle float64) *RotateEffect {
	return &RotateEffect{
		TransformEffect: TransformEffect{name: "rotate"},
		angle:           angle,
	}
}

// Apply 应用旋转特效
func (re *RotateEffect) Apply(clip core.Clip) (core.Clip, error) {
	// 这里应该返回一个新的剪辑，应用了旋转特效
	// 简化实现，直接返回原剪辑
	return clip, nil
}

// ApplyToFrame 应用旋转特效到帧
func (re *RotateEffect) ApplyToFrame(frame image.Image) (image.Image, error) {
	bounds := frame.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 检查输入尺寸是否合理
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("无效的输入尺寸: %dx%d", width, height)
	}
	if width > 8192 || height > 8192 {
		return nil, fmt.Errorf("输入尺寸过大: %dx%d", width, height)
	}

	// 将角度转换为弧度
	radians := re.angle * math.Pi / 180.0

	// 计算旋转后的尺寸
	cos := math.Cos(radians)
	sin := math.Sin(radians)
	absCos := math.Abs(cos)
	absSin := math.Abs(sin)

	// 计算旋转后的边界框（更准确的计算）
	newWidth := int(float64(width)*absCos + float64(height)*absSin)
	newHeight := int(float64(width)*absSin + float64(height)*absCos)

	// 确保尺寸是偶数（H.264编码器要求）
	if newWidth%2 != 0 {
		newWidth++
	}
	if newHeight%2 != 0 {
		newHeight++
	}

	// 限制最大尺寸，防止过大的图像
	maxDimension := 4096 // 最大4K分辨率
	if newWidth > maxDimension {
		newWidth = maxDimension
		// 确保限制后仍然是偶数
		if newWidth%2 != 0 {
			newWidth--
		}
	}
	if newHeight > maxDimension {
		newHeight = maxDimension
		// 确保限制后仍然是偶数
		if newHeight%2 != 0 {
			newHeight--
		}
	}

	// 检查计算出的尺寸是否合理
	if newWidth <= 0 || newHeight <= 0 {
		return nil, fmt.Errorf("计算出的旋转尺寸无效: %dx%d", newWidth, newHeight)
	}
	if newWidth*newHeight > 16*1024*1024 { // 限制像素总数（16M像素）
		return nil, fmt.Errorf("旋转后尺寸过大: %dx%d (%d 像素)", newWidth, newHeight, newWidth*newHeight)
	}

	// 创建新图像
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// 计算旋转中心
	centerX := float64(width) / 2.0
	centerY := float64(height) / 2.0
	newCenterX := float64(newWidth) / 2.0
	newCenterY := float64(newHeight) / 2.0

	// 应用旋转
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			// 将新坐标转换为原坐标
			dx := float64(x) - newCenterX
			dy := float64(y) - newCenterY

			// 应用逆旋转
			srcX := int(centerX + dx*cos + dy*sin)
			srcY := int(centerY - dx*sin + dy*cos)

			// 检查边界
			if srcX >= 0 && srcX < width && srcY >= 0 && srcY < height {
				dst.Set(x, y, frame.At(srcX, srcY))
			}
		}
	}

	return dst, nil
}

// CropEffect 裁剪特效
type CropEffect struct {
	TransformEffect
	x, y, width, height int
}

// NewCropEffect 创建裁剪特效
func NewCropEffect(x, y, width, height int) *CropEffect {
	return &CropEffect{
		TransformEffect: TransformEffect{name: "crop"},
		x:               x,
		y:               y,
		width:           width,
		height:          height,
	}
}

// Apply 应用裁剪特效
func (ce *CropEffect) Apply(clip core.Clip) (core.Clip, error) {
	// 这里应该返回一个新的剪辑，应用了裁剪特效
	// 简化实现，直接返回原剪辑
	return clip, nil
}

// ApplyToFrame 应用裁剪特效到帧
func (ce *CropEffect) ApplyToFrame(frame image.Image) (image.Image, error) {
	bounds := frame.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	// 确保裁剪区域在图像范围内
	if ce.x < 0 {
		ce.x = 0
	}
	if ce.y < 0 {
		ce.y = 0
	}
	if ce.x+ce.width > srcWidth {
		ce.width = srcWidth - ce.x
	}
	if ce.y+ce.height > srcHeight {
		ce.height = srcHeight - ce.y
	}

	// 创建裁剪后的图像
	dst := image.NewRGBA(image.Rect(0, 0, ce.width, ce.height))

	// 复制裁剪区域
	for y := 0; y < ce.height; y++ {
		for x := 0; x < ce.width; x++ {
			srcX := ce.x + x
			srcY := ce.y + y
			dst.Set(x, y, frame.At(srcX, srcY))
		}
	}

	return dst, nil
}

// BrightnessEffect 亮度调整特效
type BrightnessEffect struct {
	TransformEffect
	factor float64 // 亮度因子，1.0为正常，>1.0为更亮，<1.0为更暗
}

// NewBrightnessEffect 创建亮度调整特效
func NewBrightnessEffect(factor float64) *BrightnessEffect {
	return &BrightnessEffect{
		TransformEffect: TransformEffect{name: "brightness"},
		factor:          factor,
	}
}

// Apply 应用亮度调整特效
func (be *BrightnessEffect) Apply(clip core.Clip) (core.Clip, error) {
	// 这里应该返回一个新的剪辑，应用了亮度调整特效
	// 简化实现，直接返回原剪辑
	return clip, nil
}

// ApplyToFrame 应用亮度调整特效到帧
func (be *BrightnessEffect) ApplyToFrame(frame image.Image) (image.Image, error) {
	bounds := frame.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 创建新图像
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	// 应用亮度调整
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := frame.At(x, y).RGBA()

			// 调整亮度
			newR := uint32(float64(r) * be.factor)
			newG := uint32(float64(g) * be.factor)
			newB := uint32(float64(b) * be.factor)

			// 确保值在有效范围内
			if newR > 65535 {
				newR = 65535
			}
			if newG > 65535 {
				newG = 65535
			}
			if newB > 65535 {
				newB = 65535
			}

			dst.Set(x, y, color.RGBA{
				R: uint8(newR >> 8),
				G: uint8(newG >> 8),
				B: uint8(newB >> 8),
				A: uint8(a >> 8),
			})
		}
	}

	return dst, nil
}

// ContrastEffect 对比度调整特效
type ContrastEffect struct {
	TransformEffect
	factor float64 // 对比度因子，1.0为正常，>1.0为更高对比度，<1.0为更低对比度
}

// NewContrastEffect 创建对比度调整特效
func NewContrastEffect(factor float64) *ContrastEffect {
	return &ContrastEffect{
		TransformEffect: TransformEffect{name: "contrast"},
		factor:          factor,
	}
}

// Apply 应用对比度调整特效
func (ce *ContrastEffect) Apply(clip core.Clip) (core.Clip, error) {
	// 这里应该返回一个新的剪辑，应用了对比度调整特效
	// 简化实现，直接返回原剪辑
	return clip, nil
}

// ApplyToFrame 应用对比度调整特效到帧
func (ce *ContrastEffect) ApplyToFrame(frame image.Image) (image.Image, error) {
	bounds := frame.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 创建新图像
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	// 应用对比度调整
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := frame.At(x, y).RGBA()

			// 将值标准化到0-1范围
			normR := float64(r) / 65535.0
			normG := float64(g) / 65535.0
			normB := float64(b) / 65535.0

			// 应用对比度调整
			newR := (normR-0.5)*ce.factor + 0.5
			newG := (normG-0.5)*ce.factor + 0.5
			newB := (normB-0.5)*ce.factor + 0.5

			// 确保值在0-1范围内
			if newR < 0 {
				newR = 0
			} else if newR > 1 {
				newR = 1
			}
			if newG < 0 {
				newG = 0
			} else if newG > 1 {
				newG = 1
			}
			if newB < 0 {
				newB = 0
			} else if newB > 1 {
				newB = 1
			}

			// 转换回0-255范围
			dst.Set(x, y, color.RGBA{
				R: uint8(newR * 255),
				G: uint8(newG * 255),
				B: uint8(newB * 255),
				A: uint8(a >> 8),
			})
		}
	}

	return dst, nil
}
