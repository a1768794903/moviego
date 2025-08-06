package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"moviepy-go/pkg/ffmpeg"
	"moviepy-go/pkg/video"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run test_video_writing.go <视频文件>")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	fmt.Printf("测试视频写入: %s\n", inputFile)

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
	fmt.Printf("视频信息:\n")
	fmt.Printf("  文件名: %s\n", inputFile)
	fmt.Printf("  时长: %v\n", clip.Duration())
	fmt.Printf("  帧率: %.2f fps\n", clip.FPS())
	fmt.Printf("  尺寸: %dx%d\n", clip.Width(), clip.Height())

	// 测试写入短时间片段
	fmt.Printf("\n=== 测试写入短时间片段 ===\n")

	// 只处理前2秒
	duration := clip.Duration()
	if duration > 2*time.Second {
		duration = 2 * time.Second
	}

	// 创建子剪辑
	subclip, err := clip.Subclip(0, duration)
	if err != nil {
		log.Fatalf("创建子剪辑失败: %v", err)
	}
	defer subclip.Close()

	// 写入子剪辑
	outputFile := "test_output_short.mp4"
	fmt.Printf("写入短时间片段到: %s\n", outputFile)

	start := time.Now()
	if err := subclip.WriteToFile(outputFile, nil); err != nil {
		log.Fatalf("写入失败: %v", err)
	}

	elapsed := time.Since(start)
	fmt.Printf("✓ 写入完成! 耗时: %v\n", elapsed)

	// 检查输出文件
	if info, err := os.Stat(outputFile); err == nil {
		sizeMB := float64(info.Size()) / (1024 * 1024)
		fmt.Printf("输出文件大小: %.2f MB\n", sizeMB)
	} else {
		fmt.Printf("警告: 无法获取输出文件信息: %v\n", err)
	}

	fmt.Printf("\n测试完成!\n")
}
