package effects

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"moviepy-go/pkg/core"
)

// BlurEffect 模糊特效
type BlurEffect struct {
	TransformEffect
	radius int // 模糊半径
}

// Apply 应用模糊特效
func (be *BlurEffect) Apply(clip core.Clip) (core.Clip, error) {
	// 这里应该返回一个新的剪辑，应用了模糊特效
	// 简化实现，直接返回原剪辑
	return clip, nil
}

// NewBlurEffect 创建模糊特效
func NewBlurEffect(radius int) *BlurEffect {
	if radius < 1 {
		radius = 1
	}
	if radius > 20 {
		radius = 20
	}
	return &BlurEffect{
		TransformEffect: TransformEffect{name: "blur"},
		radius:          radius,
	}
}

// ApplyToFrame 应用模糊特效到帧
func (be *BlurEffect) ApplyToFrame(frame image.Image) (image.Image, error) {
	bounds := frame.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 创建新图像
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	// 应用高斯模糊
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var sumR, sumG, sumB, sumA uint32
			var count int

			// 在模糊半径内采样
			for dy := -be.radius; dy <= be.radius; dy++ {
				for dx := -be.radius; dx <= be.radius; dx++ {
					srcX := x + dx
					srcY := y + dy

					// 检查边界
					if srcX >= 0 && srcX < width && srcY >= 0 && srcY < height {
						r, g, b, a := frame.At(srcX, srcY).RGBA()
						sumR += r
						sumG += g
						sumB += b
						sumA += a
						count++
					}
				}
			}

			// 计算平均值
			if count > 0 {
				dst.Set(x, y, color.RGBA{
					R: uint8(sumR / uint32(count) >> 8),
					G: uint8(sumG / uint32(count) >> 8),
					B: uint8(sumB / uint32(count) >> 8),
					A: uint8(sumA / uint32(count) >> 8),
				})
			}
		}
	}

	return dst, nil
}

// SharpenEffect 锐化特效
type SharpenEffect struct {
	TransformEffect
	strength float64 // 锐化强度
}

// Apply 应用锐化特效
func (se *SharpenEffect) Apply(clip core.Clip) (core.Clip, error) {
	// 这里应该返回一个新的剪辑，应用了锐化特效
	// 简化实现，直接返回原剪辑
	return clip, nil
}

// NewSharpenEffect 创建锐化特效
func NewSharpenEffect(strength float64) *SharpenEffect {
	if strength < 0 {
		strength = 0
	}
	if strength > 2 {
		strength = 2
	}
	return &SharpenEffect{
		TransformEffect: TransformEffect{name: "sharpen"},
		strength:        strength,
	}
}

// ApplyToFrame 应用锐化特效到帧
func (se *SharpenEffect) ApplyToFrame(frame image.Image) (image.Image, error) {
	bounds := frame.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 创建新图像
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	// 锐化卷积核
	kernel := [3][3]float64{
		{0, -1, 0},
		{-1, 5, -1},
		{0, -1, 0},
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var sumR, sumG, sumB, sumA float64

			// 应用卷积核
			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
					srcX := x + kx
					srcY := y + ky

					// 边界处理
					if srcX < 0 {
						srcX = 0
					} else if srcX >= width {
						srcX = width - 1
					}
					if srcY < 0 {
						srcY = 0
					} else if srcY >= height {
						srcY = height - 1
					}

					r, g, b, a := frame.At(srcX, srcY).RGBA()
					weight := kernel[ky+1][kx+1] * se.strength

					sumR += float64(r) * weight
					sumG += float64(g) * weight
					sumB += float64(b) * weight
					sumA += float64(a) * weight
				}
			}

			// 确保值在有效范围内
			r := int(sumR)
			g := int(sumG)
			b := int(sumB)
			a := int(sumA)

			if r < 0 {
				r = 0
			} else if r > 65535 {
				r = 65535
			}
			if g < 0 {
				g = 0
			} else if g > 65535 {
				g = 65535
			}
			if b < 0 {
				b = 0
			} else if b > 65535 {
				b = 65535
			}
			if a < 0 {
				a = 0
			} else if a > 65535 {
				a = 65535
			}

			dst.Set(x, y, color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			})
		}
	}

	return dst, nil
}

// SaturationEffect 饱和度调整特效
type SaturationEffect struct {
	TransformEffect
	factor float64 // 饱和度因子，1.0为正常，>1.0为更高饱和度，<1.0为更低饱和度
}

// Apply 应用饱和度调整特效
func (se *SaturationEffect) Apply(clip core.Clip) (core.Clip, error) {
	// 这里应该返回一个新的剪辑，应用了饱和度调整特效
	// 简化实现，直接返回原剪辑
	return clip, nil
}

// NewSaturationEffect 创建饱和度调整特效
func NewSaturationEffect(factor float64) *SaturationEffect {
	return &SaturationEffect{
		TransformEffect: TransformEffect{name: "saturation"},
		factor:          factor,
	}
}

// ApplyToFrame 应用饱和度调整特效到帧
func (se *SaturationEffect) ApplyToFrame(frame image.Image) (image.Image, error) {
	bounds := frame.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 创建新图像
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := frame.At(x, y).RGBA()

			// 转换为HSL
			rf := float64(r) / 65535.0
			gf := float64(g) / 65535.0
			bf := float64(b) / 65535.0

			// 计算亮度
			luminance := 0.299*rf + 0.587*gf + 0.114*bf

			// 调整饱和度
			newR := luminance + (rf-luminance)*se.factor
			newG := luminance + (gf-luminance)*se.factor
			newB := luminance + (bf-luminance)*se.factor

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

// NoiseEffect 噪点特效
type NoiseEffect struct {
	TransformEffect
	intensity float64 // 噪点强度，0.0为无噪点，1.0为最大噪点
}

// Apply 应用噪点特效
func (ne *NoiseEffect) Apply(clip core.Clip) (core.Clip, error) {
	// 这里应该返回一个新的剪辑，应用了噪点特效
	// 简化实现，直接返回原剪辑
	return clip, nil
}

// NewNoiseEffect 创建噪点特效
func NewNoiseEffect(intensity float64) *NoiseEffect {
	if intensity < 0 {
		intensity = 0
	}
	if intensity > 1 {
		intensity = 1
	}
	return &NoiseEffect{
		TransformEffect: TransformEffect{name: "noise"},
		intensity:       intensity,
	}
}

// ApplyToFrame 应用噪点特效到帧
func (ne *NoiseEffect) ApplyToFrame(frame image.Image) (image.Image, error) {
	bounds := frame.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 创建新图像
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := frame.At(x, y).RGBA()

			// 生成随机噪点
			noise := (rand.Float64() - 0.5) * 2 * ne.intensity

			// 应用噪点
			newR := float64(r)/65535.0 + noise
			newG := float64(g)/65535.0 + noise
			newB := float64(b)/65535.0 + noise

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

// SepiaEffect 棕褐色特效
type SepiaEffect struct {
	TransformEffect
	strength float64 // 棕褐色强度，0.0为原色，1.0为完全棕褐色
}

// Apply 应用棕褐色特效
func (se *SepiaEffect) Apply(clip core.Clip) (core.Clip, error) {
	// 这里应该返回一个新的剪辑，应用了棕褐色特效
	// 简化实现，直接返回原剪辑
	return clip, nil
}

// NewSepiaEffect 创建棕褐色特效
func NewSepiaEffect(strength float64) *SepiaEffect {
	if strength < 0 {
		strength = 0
	}
	if strength > 1 {
		strength = 1
	}
	return &SepiaEffect{
		TransformEffect: TransformEffect{name: "sepia"},
		strength:        strength,
	}
}

// ApplyToFrame 应用棕褐色特效到帧
func (se *SepiaEffect) ApplyToFrame(frame image.Image) (image.Image, error) {
	bounds := frame.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 创建新图像
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := frame.At(x, y).RGBA()

			// 转换为灰度
			gray := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)

			// 应用棕褐色滤镜
			sepiaR := gray*0.393 + gray*0.769 + gray*0.189
			sepiaG := gray*0.349 + gray*0.686 + gray*0.168
			sepiaB := gray*0.272 + gray*0.534 + gray*0.131

			// 混合原色和棕褐色
			finalR := float64(r)*(1-se.strength) + sepiaR*se.strength
			finalG := float64(g)*(1-se.strength) + sepiaG*se.strength
			finalB := float64(b)*(1-se.strength) + sepiaB*se.strength

			// 确保值在有效范围内
			if finalR > 65535 {
				finalR = 65535
			}
			if finalG > 65535 {
				finalG = 65535
			}
			if finalB > 65535 {
				finalB = 65535
			}

			dst.Set(x, y, color.RGBA{
				R: uint8(finalR / 256),
				G: uint8(finalG / 256),
				B: uint8(finalB / 256),
				A: uint8(a >> 8),
			})
		}
	}

	return dst, nil
}

// VignetteEffect 暗角特效
type VignetteEffect struct {
	TransformEffect
	strength float64 // 暗角强度，0.0为无暗角，1.0为最强暗角
	radius   float64 // 暗角半径，0.0为中心点，1.0为整个图像
}

// Apply 应用暗角特效
func (ve *VignetteEffect) Apply(clip core.Clip) (core.Clip, error) {
	// 这里应该返回一个新的剪辑，应用了暗角特效
	// 简化实现，直接返回原剪辑
	return clip, nil
}

// NewVignetteEffect 创建暗角特效
func NewVignetteEffect(strength, radius float64) *VignetteEffect {
	if strength < 0 {
		strength = 0
	}
	if strength > 1 {
		strength = 1
	}
	if radius < 0 {
		radius = 0
	}
	if radius > 1 {
		radius = 1
	}
	return &VignetteEffect{
		TransformEffect: TransformEffect{name: "vignette"},
		strength:        strength,
		radius:          radius,
	}
}

// ApplyToFrame 应用暗角特效到帧
func (ve *VignetteEffect) ApplyToFrame(frame image.Image) (image.Image, error) {
	bounds := frame.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 创建新图像
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	// 计算中心点
	centerX := float64(width) / 2.0
	centerY := float64(height) / 2.0
	maxDistance := math.Sqrt(centerX*centerX+centerY*centerY) * ve.radius

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := frame.At(x, y).RGBA()

			// 计算到中心的距离
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			distance := math.Sqrt(dx*dx + dy*dy)

			// 计算暗角因子
			vignetteFactor := 1.0
			if distance > 0 {
				vignetteFactor = 1.0 - (distance/maxDistance)*ve.strength
				if vignetteFactor < 0 {
					vignetteFactor = 0
				}
			}

			// 应用暗角
			newR := uint32(float64(r) * vignetteFactor)
			newG := uint32(float64(g) * vignetteFactor)
			newB := uint32(float64(b) * vignetteFactor)

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
