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
		fmt.Println("用法: go run simple_test.go <视频文件>")
		os.Exit(1)
	}

	filename := os.Args[1]
	fmt.Printf("测试文件: %s\n", filename)

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

	// 测试获取第一帧
	fmt.Println("\n测试获取第一帧...")
	start := time.Now()

	frame, err := reader.GetFrame(0)
	if err != nil {
		log.Fatalf("获取帧失败: %v", err)
	}

	duration := time.Since(start)
	bounds := frame.Bounds()

	fmt.Printf("成功! 耗时: %v, 帧尺寸: %dx%d\n", duration, bounds.Dx(), bounds.Dy())

	fmt.Println("测试完成!")
}
