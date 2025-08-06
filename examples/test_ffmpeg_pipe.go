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
		fmt.Println("用法: go run test_ffmpeg_pipe.go <视频文件>")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	fmt.Printf("测试FFmpeg管道稳定性: %s\n", inputFile)

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
	fmt.Printf("  文件名: %s\n", inputFile)
	fmt.Printf("  时长: %v\n", clip.Duration())
	fmt.Printf("  帧率: %.2f fps\n", clip.FPS())
	fmt.Printf("  尺寸: %dx%d\n", clip.Width(), clip.Height())

	// 创建组合特效剪辑
	fmt.Printf("\n=== 创建组合特效剪辑 ===\n")

	effectClip := video.NewEffectVideoClip(clip, processMgr)

	// 添加缩放特效
	effectClip.AddEffect(effects.NewResizeEffect(480, 360))
	fmt.Printf("添加缩放特效后尺寸: %dx%d\n", effectClip.Width(), effectClip.Height())

	// 添加旋转特效
	effectClip.AddEffect(effects.NewRotateEffect(90.0))
	fmt.Printf("添加旋转特效后尺寸: %dx%d\n", effectClip.Width(), effectClip.Height())

	// 添加亮度特效
	effectClip.AddEffect(effects.NewBrightnessEffect(1.3))
	fmt.Printf("添加亮度特效后尺寸: %dx%d\n", effectClip.Width(), effectClip.Height())

	// 创建短剪辑（前2秒）
	duration := effectClip.Duration()
	if duration > 2*time.Second {
		duration = 2 * time.Second
	}

	subclip, err := effectClip.Subclip(0, duration)
	if err != nil {
		log.Fatalf("创建子剪辑失败: %v", err)
	}
	defer subclip.Close()

	// 测试逐帧获取
	fmt.Printf("\n=== 测试帧获取 ===\n")

	frameInterval := time.Duration(float64(time.Second) / effectClip.FPS())
	frameCount := int(duration.Seconds() * effectClip.FPS())

	fmt.Printf("预期帧数: %d\n", frameCount)
	fmt.Printf("帧间隔: %v\n", frameInterval)

	for i := 0; i < min(10, frameCount); i++ { // 只测试前10帧
		t := time.Duration(i) * frameInterval
		frame, err := subclip.GetFrame(t)
		if err != nil {
			log.Printf("❌ 获取第 %d 帧失败: %v", i, err)
			break
		}
		bounds := frame.Bounds()
		fmt.Printf("第 %d 帧: %dx%d (时间: %v)\n", i, bounds.Dx(), bounds.Dy(), t)
	}

	// 设置编码选项
	options := &core.WriteOptions{
		Codec:   "libx264",
		Bitrate: "1000k",
		FPS:     effectClip.FPS(),
	}

	fmt.Printf("\n=== 测试视频写入 ===\n")
	fmt.Printf("写入组合特效视频到: test_ffmpeg_pipe.mp4\n")

	// 写入视频
	start := time.Now()
	if err := subclip.WriteToFile("test_ffmpeg_pipe.mp4", options); err != nil {
		log.Printf("❌ 写入失败: %v", err)

		// 尝试获取更多诊断信息
		effectClip.Close()
		return
	}

	elapsed := time.Since(start)
	fmt.Printf("✓ 写入完成! 耗时: %v\n", elapsed)

	// 检查输出文件
	if info, err := os.Stat("test_ffmpeg_pipe.mp4"); err == nil {
		sizeMB := float64(info.Size()) / (1024 * 1024)
		fmt.Printf("文件大小: %.2f MB\n", sizeMB)
	} else {
		fmt.Printf("警告: 无法获取文件信息: %v\n", err)
	}

	effectClip.Close()

	fmt.Printf("\n测试完成!\n")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
