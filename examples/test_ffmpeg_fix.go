package main

import (
	"fmt"
	"log"
	"time"

	"moviepy-go/pkg/ffmpeg"
	"moviepy-go/pkg/video"
)

func main() {
	fmt.Println("测试 FFmpeg 进程管理修复...")

	// 创建进程管理器
	processMgr := ffmpeg.NewProcessManager()
	defer processMgr.Close()

	// 创建视频文件剪辑
	clip := video.NewVideoFileClip("../IMG_8743.MP4", processMgr)

	// 打开视频
	if err := clip.Open(); err != nil {
		log.Fatalf("打开视频失败: %v", err)
	}
	defer clip.Close()

	fmt.Printf("视频信息: 时长=%.2fs, 尺寸=%dx%d, FPS=%.2f\n",
		clip.Duration().Seconds(), clip.Width(), clip.Height(), clip.FPS())

	// 测试获取多个帧
	frameTimes := []time.Duration{
		0 * time.Second,
		1 * time.Second,
		2 * time.Second,
		3 * time.Second,
	}

	for i, t := range frameTimes {
		fmt.Printf("获取第 %d 帧 (时间: %v)...\n", i+1, t)

		frame, err := clip.GetFrame(t)
		if err != nil {
			log.Printf("获取帧失败 (t=%v): %v", t, err)
			continue
		}

		bounds := frame.Bounds()
		fmt.Printf("  帧尺寸: %dx%d\n", bounds.Dx(), bounds.Dy())
	}

	fmt.Println("测试完成！")
}
