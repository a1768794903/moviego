package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"moviepy-go/pkg/core"
	"moviepy-go/pkg/effects"
	"moviepy-go/pkg/ffmpeg"
	"moviepy-go/pkg/video"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run test_effects_dimensions.go <视频文件>")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	fmt.Printf("测试特效尺寸计算: %s\n", inputFile)

	// 创建进程管理器
	processMgr := ffmpeg.NewProcessManager()
	defer processMgr.Close()

	// 创建视频剪辑
	clip := video.NewVideoFileClip(inputFile, processMgr)

	// 打开视频文件
	fmt.Println("正在打开视频...")
	if err := clip.Open(); err != nil {
		log.Fatalf("打开视频失败: %v", err)
	}
	defer clip.Close()

	// 打印视频信息
	fmt.Printf("原始视频信息:\n")
	fmt.Printf("  尺寸: %dx%d\n", clip.Width(), clip.Height())

	// 测试不同特效的尺寸计算
	testCases := []struct {
		name   string
		effect effects.VideoEffect
	}{
		{
			name:   "缩放到640x480",
			effect: effects.NewResizeEffect(640, 480),
		},
		{
			name:   "旋转90度",
			effect: effects.NewRotateEffect(90.0),
		},
		{
			name:   "旋转45度",
			effect: effects.NewRotateEffect(45.0),
		},
		{
			name:   "裁剪中心区域",
			effect: effects.NewCropEffect(clip.Width()/4, clip.Height()/4, clip.Width()/2, clip.Height()/2),
		},
		{
			name:   "亮度调整(不改变尺寸)",
			effect: effects.NewBrightnessEffect(1.5),
		},
	}

	for _, testCase := range testCases {
		fmt.Printf("\n=== 测试 %s ===\n", testCase.name)

		// 创建特效剪辑
		effectClip := video.NewEffectVideoClip(clip, processMgr)

		fmt.Printf("添加特效前尺寸: %dx%d\n", effectClip.Width(), effectClip.Height())

		// 添加特效
		effectClip.AddEffect(testCase.effect)

		fmt.Printf("添加特效后尺寸: %dx%d\n", effectClip.Width(), effectClip.Height())

		// 测试获取帧
		frame, err := effectClip.GetFrame(0)
		if err != nil {
			log.Printf("❌ 获取帧失败: %v", err)
			effectClip.Close()
			continue
		}

		bounds := frame.Bounds()
		actualWidth := bounds.Dx()
		actualHeight := bounds.Dy()

		fmt.Printf("实际帧尺寸: %dx%d\n", actualWidth, actualHeight)

		// 检查尺寸是否匹配
		if actualWidth == effectClip.Width() && actualHeight == effectClip.Height() {
			fmt.Printf("✓ 尺寸匹配正确\n")
		} else {
			fmt.Printf("❌ 尺寸不匹配！期望: %dx%d, 实际: %dx%d\n",
				effectClip.Width(), effectClip.Height(), actualWidth, actualHeight)
		}

		effectClip.Close()
	}

	// 测试组合特效
	fmt.Printf("\n=== 测试组合特效尺寸计算 ===\n")

	comboClip := video.NewEffectVideoClip(clip, processMgr)
	fmt.Printf("初始尺寸: %dx%d\n", comboClip.Width(), comboClip.Height())

	// 添加缩放特效
	comboClip.AddEffect(effects.NewResizeEffect(800, 600))
	fmt.Printf("添加缩放特效后: %dx%d\n", comboClip.Width(), comboClip.Height())

	// 添加旋转特效
	comboClip.AddEffect(effects.NewRotateEffect(90.0))
	fmt.Printf("添加旋转特效后: %dx%d\n", comboClip.Width(), comboClip.Height())

	// 添加裁剪特效
	comboClip.AddEffect(effects.NewCropEffect(50, 50, 400, 300))
	fmt.Printf("添加裁剪特效后: %dx%d\n", comboClip.Width(), comboClip.Height())

	// 测试获取帧
	frame, err := comboClip.GetFrame(0)
	if err != nil {
		log.Printf("❌ 获取组合特效帧失败: %v", err)
	} else {
		bounds := frame.Bounds()
		actualWidth := bounds.Dx()
		actualHeight := bounds.Dy()

		fmt.Printf("组合特效实际帧尺寸: %dx%d\n", actualWidth, actualHeight)

		if actualWidth == comboClip.Width() && actualHeight == comboClip.Height() {
			fmt.Printf("✓ 组合特效尺寸匹配正确\n")
		} else {
			fmt.Printf("❌ 组合特效尺寸不匹配！期望: %dx%d, 实际: %dx%d\n",
				comboClip.Width(), comboClip.Height(), actualWidth, actualHeight)
		}
	}

	comboClip.Close()

	// 测试写入
	fmt.Printf("\n=== 测试写入组合特效 ===\n")

	writeClip := video.NewEffectVideoClip(clip, processMgr)
	writeClip.AddEffect(effects.NewResizeEffect(640, 480))
	writeClip.AddEffect(effects.NewRotateEffect(45.0))

	fmt.Printf("写入测试剪辑尺寸: %dx%d\n", writeClip.Width(), writeClip.Height())

	// 创建短剪辑
	duration := writeClip.Duration()
	if duration > 1*time.Second {
		duration = 1 * time.Second
	}

	subclip, err := writeClip.Subclip(0, duration)
	if err != nil {
		log.Fatalf("创建子剪辑失败: %v", err)
	}
	defer subclip.Close()

	// 设置编码选项
	options := &core.WriteOptions{
		Codec:   "libx264",
		Bitrate: "1000k",
		FPS:     writeClip.FPS(),
	}

	fmt.Printf("写入测试视频到: test_effects_dimensions.mp4\n")

	// 写入视频
	start := time.Now()
	if err := subclip.WriteToFile("test_effects_dimensions.mp4", options); err != nil {
		log.Printf("❌ 写入失败: %v", err)
	} else {
		elapsed := time.Since(start)
		fmt.Printf("✓ 写入完成! 耗时: %v\n", elapsed)
	}

	writeClip.Close()

	fmt.Printf("\n测试完成!\n")
}
