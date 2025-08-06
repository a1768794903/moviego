package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"moviepy-go/pkg/ffmpeg"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run debug_ffmpeg.go <视频文件>")
		os.Exit(1)
	}

	filename := os.Args[1]
	fmt.Printf("测试 FFmpeg 读取器: %s\n", filename)

	// 创建进程管理器
	processMgr := ffmpeg.NewProcessManager()
	defer processMgr.Close()

	// 创建视频读取器
	reader := ffmpeg.NewVideoReader(filename, processMgr)

	// 打开视频
	fmt.Println("正在打开视频...")
	if err := reader.Open(); err != nil {
		log.Fatalf("打开视频失败: %v", err)
	}
	defer reader.Close()

	// 获取视频信息
	info := reader.GetInfo()
	if info == nil {
		log.Fatal("无法获取视频信息")
	}

	fmt.Printf("视频信息:\n")
	fmt.Printf("  时长: %.2f 秒\n", info.Duration)
	fmt.Printf("  尺寸: %dx%d\n", info.Width, info.Height)
	fmt.Printf("  帧率: %.2f fps\n", info.FPS)
	fmt.Printf("  有音频: %v\n", info.HasAudio)

	// 测试获取第一帧
	fmt.Println("\n测试获取第一帧...")
	start := time.Now()

	frame, err := reader.GetFrame(0)
	if err != nil {
		log.Fatalf("获取帧失败: %v", err)
	}

	duration := time.Since(start)
	bounds := frame.Bounds()

	fmt.Printf("成功获取帧:\n")
	fmt.Printf("  耗时: %v\n", duration)
	fmt.Printf("  帧尺寸: %dx%d\n", bounds.Dx(), bounds.Dy())

	// 测试获取第二帧
	fmt.Println("\n测试获取第二帧...")
	start = time.Now()

	frame2, err := reader.GetFrame(time.Second)
	if err != nil {
		log.Printf("获取第二帧失败: %v", err)
	} else {
		duration = time.Since(start)
		bounds = frame2.Bounds()
		fmt.Printf("成功获取第二帧:\n")
		fmt.Printf("  耗时: %v\n", duration)
		fmt.Printf("  帧尺寸: %dx%d\n", bounds.Dx(), bounds.Dy())
	}

	fmt.Println("\n测试完成!")
}
