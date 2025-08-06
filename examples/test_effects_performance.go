package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"moviepy-go/pkg/effects"
	"moviepy-go/pkg/ffmpeg"
	"moviepy-go/pkg/video"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run test_effects_performance.go <视频文件>")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	fmt.Printf("测试特效性能: %s\n", inputFile)

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
	fmt.Printf("  帧率: %.2f fps\n", clip.FPS())

	// 测试单个特效的性能
	testEffects := []struct {
		name   string
		effect effects.VideoEffect
	}{
		{
			name:   "缩放到480x360",
			effect: effects.NewResizeEffect(480, 360),
		},
		{
			name:   "旋转90度",
			effect: effects.NewRotateEffect(90.0),
		},
		{
			name:   "亮度调整",
			effect: effects.NewBrightnessEffect(1.3),
		},
	}

	// 获取原始帧
	fmt.Println("\n=== 获取原始帧 ===")
	start := time.Now()
	originalFrame, err := clip.GetFrame(0)
	if err != nil {
		log.Fatalf("获取原始帧失败: %v", err)
	}
	elapsed := time.Since(start)
	bounds := originalFrame.Bounds()
	fmt.Printf("✓ 原始帧获取耗时: %v, 尺寸: %dx%d\n", elapsed, bounds.Dx(), bounds.Dy())

	// 测试每个特效的性能
	for _, test := range testEffects {
		fmt.Printf("\n=== 测试 %s ===\n", test.name)

		start := time.Now()
		resultFrame, err := test.effect.ApplyToFrame(originalFrame)
		elapsed := time.Since(start)

		if err != nil {
			log.Printf("❌ %s 失败: %v", test.name, err)
			continue
		}

		bounds := resultFrame.Bounds()
		fmt.Printf("✓ %s 耗时: %v, 结果尺寸: %dx%d\n", test.name, elapsed, bounds.Dx(), bounds.Dy())
	}

	// 测试组合特效的性能
	fmt.Printf("\n=== 测试组合特效 ===\n")

	// 创建组合特效剪辑
	effectClip := video.NewEffectVideoClip(clip, processMgr)

	// 逐步添加特效并测试性能
	steps := []struct {
		effect effects.VideoEffect
		name   string
	}{
		{effects.NewResizeEffect(480, 360), "缩放"},
		{effects.NewRotateEffect(90.0), "旋转"},
		{effects.NewBrightnessEffect(1.3), "亮度"},
	}

	for i, step := range steps {
		fmt.Printf("\n--- 添加 %s 特效 ---\n", step.name)

		effectClip.AddEffect(step.effect)
		fmt.Printf("当前剪辑尺寸: %dx%d\n", effectClip.Width(), effectClip.Height())

		// 测试获取帧的性能
		start := time.Now()
		frame, err := effectClip.GetFrame(0)
		elapsed := time.Since(start)

		if err != nil {
			log.Printf("❌ 获取组合特效帧失败 (步骤 %d): %v", i+1, err)
			break
		}

		bounds := frame.Bounds()
		fmt.Printf("✓ 组合特效帧获取耗时: %v, 尺寸: %dx%d\n", elapsed, bounds.Dx(), bounds.Dy())

		if elapsed > 5*time.Second {
			fmt.Printf("⚠️ 性能警告: 帧处理时间过长 (%v)\n", elapsed)
		}
	}

	// 测试多帧获取
	fmt.Printf("\n=== 测试多帧获取 ===\n")

	frameInterval := time.Duration(float64(time.Second) / effectClip.FPS())

	for i := 0; i < 3; i++ {
		t := time.Duration(i) * frameInterval

		fmt.Printf("获取第 %d 帧 (时间: %v)...", i, t)
		start := time.Now()
		frame, err := effectClip.GetFrame(t)
		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf(" ❌ 失败: %v\n", err)
			break
		}

		bounds := frame.Bounds()
		fmt.Printf(" ✓ 耗时: %v, 尺寸: %dx%d\n", elapsed, bounds.Dx(), bounds.Dy())

		if elapsed > 5*time.Second {
			fmt.Printf("⚠️ 第 %d 帧处理时间过长，停止测试\n", i)
			break
		}
	}

	effectClip.Close()

	fmt.Printf("\n测试完成!\n")
}
