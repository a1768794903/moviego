package main

import (
	"fmt"
	"image"
	"image/color"

	"moviepy-go/pkg/effects"
)

func main() {
	fmt.Println("=== 特效系统基础测试 ===")

	// 创建一个测试图像
	testImage := createTestImage(100, 100)
	fmt.Printf("创建测试图像: %dx%d\n", testImage.Bounds().Dx(), testImage.Bounds().Dy())

	// 测试基础特效
	testBasicEffects(testImage)

	// 测试高级特效
	testAdvancedEffects(testImage)

	// 测试特效链
	testEffectChains(testImage)

	fmt.Println("\n特效系统基础测试完成!")
}

// createTestImage 创建测试图像
func createTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 创建渐变图像
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8(float64(x) / float64(width) * 255)
			g := uint8(float64(y) / float64(height) * 255)
			b := uint8((float64(x) + float64(y)) / float64(width+height) * 255)

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	return img
}

// testBasicEffects 测试基础特效
func testBasicEffects(img image.Image) {
	fmt.Printf("\n--- 基础特效测试 ---\n")

	// 测试亮度调整
	fmt.Printf("1. 亮度调整特效\n")
	brightnessEffect := effects.NewBrightnessEffect(1.5)
	result, err := brightnessEffect.ApplyToFrame(img)
	if err != nil {
		fmt.Printf("  错误: %v\n", err)
	} else {
		fmt.Printf("  成功应用亮度特效\n")
		_ = result // 避免未使用变量警告
	}

	// 测试对比度调整
	fmt.Printf("2. 对比度调整特效\n")
	contrastEffect := effects.NewContrastEffect(1.3)
	result, err = contrastEffect.ApplyToFrame(img)
	if err != nil {
		fmt.Printf("  错误: %v\n", err)
	} else {
		fmt.Printf("  成功应用对比度特效\n")
		_ = result
	}

	// 测试缩放特效
	fmt.Printf("3. 缩放特效\n")
	resizeEffect := effects.NewResizeEffect(200, 150)
	result, err = resizeEffect.ApplyToFrame(img)
	if err != nil {
		fmt.Printf("  错误: %v\n", err)
	} else {
		fmt.Printf("  成功应用缩放特效\n")
		_ = result
	}

	// 测试旋转特效
	fmt.Printf("4. 旋转特效\n")
	rotateEffect := effects.NewRotateEffect(45.0)
	result, err = rotateEffect.ApplyToFrame(img)
	if err != nil {
		fmt.Printf("  错误: %v\n", err)
	} else {
		fmt.Printf("  成功应用旋转特效\n")
		_ = result
	}

	// 测试裁剪特效
	fmt.Printf("5. 裁剪特效\n")
	cropEffect := effects.NewCropEffect(25, 25, 50, 50)
	result, err = cropEffect.ApplyToFrame(img)
	if err != nil {
		fmt.Printf("  错误: %v\n", err)
	} else {
		fmt.Printf("  成功应用裁剪特效\n")
		_ = result
	}
}

// testAdvancedEffects 测试高级特效
func testAdvancedEffects(img image.Image) {
	fmt.Printf("\n--- 高级特效测试 ---\n")

	// 测试模糊特效
	fmt.Printf("1. 模糊特效\n")
	blurEffect := effects.NewBlurEffect(3)
	result, err := blurEffect.ApplyToFrame(img)
	if err != nil {
		fmt.Printf("  错误: %v\n", err)
	} else {
		fmt.Printf("  成功应用模糊特效\n")
		_ = result
	}

	// 测试锐化特效
	fmt.Printf("2. 锐化特效\n")
	sharpenEffect := effects.NewSharpenEffect(1.2)
	result, err = sharpenEffect.ApplyToFrame(img)
	if err != nil {
		fmt.Printf("  错误: %v\n", err)
	} else {
		fmt.Printf("  成功应用锐化特效\n")
		_ = result
	}

	// 测试饱和度调整
	fmt.Printf("3. 饱和度调整特效\n")
	saturationEffect := effects.NewSaturationEffect(1.5)
	result, err = saturationEffect.ApplyToFrame(img)
	if err != nil {
		fmt.Printf("  错误: %v\n", err)
	} else {
		fmt.Printf("  成功应用饱和度调整特效\n")
		_ = result
	}

	// 测试棕褐色特效
	fmt.Printf("4. 棕褐色特效\n")
	sepiaEffect := effects.NewSepiaEffect(0.7)
	result, err = sepiaEffect.ApplyToFrame(img)
	if err != nil {
		fmt.Printf("  错误: %v\n", err)
	} else {
		fmt.Printf("  成功应用棕褐色特效\n")
		_ = result
	}

	// 测试暗角特效
	fmt.Printf("5. 暗角特效\n")
	vignetteEffect := effects.NewVignetteEffect(0.4, 0.8)
	result, err = vignetteEffect.ApplyToFrame(img)
	if err != nil {
		fmt.Printf("  错误: %v\n", err)
	} else {
		fmt.Printf("  成功应用暗角特效\n")
		_ = result
	}
}

// testEffectChains 测试特效链
func testEffectChains(img image.Image) {
	fmt.Printf("\n--- 特效链测试 ---\n")

	// 测试特效构建器
	fmt.Printf("1. 特效构建器\n")
	builder := effects.NewEffectBuilder()
	chain := builder.
		Brightness(1.2).
		Contrast(1.1).
		Saturation(1.3).
		Build()
	fmt.Printf("  创建特效链，包含 %d 个特效\n", len(chain.GetEffects()))

	// 测试特效链应用
	result, err := chain.ApplyToFrame(img)
	if err != nil {
		fmt.Printf("  应用特效链错误: %v\n", err)
	} else {
		fmt.Printf("  成功应用特效链\n")
		_ = result
	}

	// 测试特效预设
	fmt.Printf("2. 特效预设\n")

	presets := []struct {
		name  string
		chain *effects.EffectChain
	}{
		{"复古", effects.Vintage()},
		{"电影", effects.Cinematic()},
		{"暖色调", effects.Warm()},
		{"冷色调", effects.Cool()},
		{"戏剧性", effects.Dramatic()},
	}

	for _, preset := range presets {
		fmt.Printf("  测试 %s 预设: %d 个特效\n", preset.name, len(preset.chain.GetEffects()))
	}
}
