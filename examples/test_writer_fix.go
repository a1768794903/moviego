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
		fmt.Println("用法: go run test_writer_fix.go <视频文件>")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	fmt.Printf("测试VideoWriter修复: %s\n", inputFile)

	// 创建进程管理器
	processMgr := ffmpeg.NewProcessManager()
	defer processMgr.Close()

	// 创建视频剪辑
	clip := video.NewVideoFileClip(inputFile, processMgr)

	// 打开视频文件
	if err := clip.Open(); err != nil {
		log.Fatalf("打开视频失败: %v", err)
	}
	defer clip.Close()

	fmt.Printf("视频信息: %dx%d @ %.2f fps\n", clip.Width(), clip.Height(), clip.FPS())

	// 创建简单的组合特效剪辑
	effectClip := video.NewEffectVideoClip(clip, processMgr)
	defer effectClip.Close()

	// 添加特效（使用相对保守的设置）
	effectClip.AddEffect(effects.NewResizeEffect(640, 480)) // 适中的尺寸
	effectClip.AddEffect(effects.NewRotateEffect(45.0))     // 45度旋转，比90度更简单
	effectClip.AddEffect(effects.NewBrightnessEffect(1.1))  // 轻微亮度调整

	fmt.Printf("特效剪辑尺寸: %dx%d\n", effectClip.Width(), effectClip.Height())

	// 测试单帧获取
	fmt.Println("\n=== 测试单帧获取 ===")
	for i := 0; i < 3; i++ {
		frameTime := time.Duration(i) * time.Duration(float64(time.Second)/effectClip.FPS())

		fmt.Printf("获取第 %d 帧 (时间: %v)...", i, frameTime)
		start := time.Now()

		frame, err := effectClip.GetFrame(frameTime)
		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf(" ❌ 失败: %v\n", err)
			return
		}

		bounds := frame.Bounds()
		fmt.Printf(" ✓ 成功，耗时: %v, 尺寸: %dx%d\n", elapsed, bounds.Dx(), bounds.Dy())

		if elapsed > 10*time.Second {
			fmt.Printf("⚠️ 第 %d 帧处理时间过长，停止测试\n", i)
			return
		}
	}

	// 测试写入（只写入前3帧）
	fmt.Println("\n=== 测试视频写入 ===")
	outputFile := "test_writer_fix.mp4"

	writer := ffmpeg.NewVideoWriter(outputFile, effectClip.Width(), effectClip.Height(), &ffmpeg.VideoWriterOptions{
		Codec:   "libx264",
		Bitrate: "1000k",
	}, processMgr)

	fmt.Printf("创建写入器: %s (%dx%d @ %.2f fps)\n", outputFile, effectClip.Width(), effectClip.Height(), effectClip.FPS())

	start := time.Now()
	if err := writer.Open(); err != nil {
		log.Fatalf("打开写入器失败: %v", err)
	}
	elapsed := time.Since(start)
	fmt.Printf("✓ 写入器打开耗时: %v\n", elapsed)

	// 写入前3帧
	frameInterval := time.Duration(float64(time.Second) / effectClip.FPS())

	for i := 0; i < 3; i++ {
		frameTime := time.Duration(i) * frameInterval

		fmt.Printf("写入第 %d 帧...", i)
		start = time.Now()

		// 获取帧
		frame, err := effectClip.GetFrame(frameTime)
		if err != nil {
			fmt.Printf(" ❌ 获取帧失败: %v\n", err)
			break
		}

		getElapsed := time.Since(start)

		// 写入帧
		writeStart := time.Now()
		if err := writer.WriteFrame(frame); err != nil {
			fmt.Printf(" ❌ 写入失败: %v\n", err)
			break
		}
		writeElapsed := time.Since(writeStart)

		totalElapsed := time.Since(start)
		fmt.Printf(" ✓ 成功，获取: %v, 写入: %v, 总计: %v\n", getElapsed, writeElapsed, totalElapsed)

		if totalElapsed > 10*time.Second {
			fmt.Printf("⚠️ 第 %d 帧处理时间过长，停止测试\n", i)
			break
		}
	}

	// 关闭写入器
	fmt.Println("关闭写入器...")
	start = time.Now()
	if err := writer.Close(); err != nil {
		log.Fatalf("关闭写入器失败: %v", err)
	}
	elapsed = time.Since(start)
	fmt.Printf("✓ 写入器关闭耗时: %v\n", elapsed)

	fmt.Printf("\n测试完成! 输出文件: %s\n", outputFile)
}
