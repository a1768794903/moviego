package main

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"moviepy-go/pkg/core"
	"moviepy-go/pkg/ffmpeg"
	"moviepy-go/pkg/video"
)

func main() {
	// 创建进程管理器
	processMgr := ffmpeg.NewProcessManager()
	defer processMgr.Close()

	// 创建视频剪辑
	clip := video.NewVideoFileClip("test_video.mp4", processMgr)

	// 打开视频文件
	if err := clip.Open(); err != nil {
		log.Printf("打开视频失败: %v", err)
		log.Println("创建测试视频...")

		// 创建一个简单的测试视频
		createTestVideo(processMgr)
		return
	}
	defer clip.Close()

	// 设置写入选项
	options := &core.WriteOptions{
		Codec:   "libx264",
		Bitrate: "1000k",
		FPS:     25.0,
	}

	// 写入视频文件
	fmt.Println("开始写入视频...")
	if err := clip.WriteToFile("output.mp4", options); err != nil {
		log.Fatalf("写入视频失败: %v", err)
	}

	fmt.Println("视频写入完成!")
}

// createTestVideo 创建一个简单的测试视频
func createTestVideo(processMgr *ffmpeg.ProcessManager) {
	// 创建视频写入器
	writerOptions := &ffmpeg.VideoWriterOptions{
		Codec:   "libx264",
		Bitrate: "1000k",
		FPS:     25.0,
	}

	writer := ffmpeg.NewVideoWriter("test_video.mp4", 640, 480, writerOptions, processMgr)

	// 打开写入器
	if err := writer.Open(); err != nil {
		log.Fatalf("打开写入器失败: %v", err)
	}
	defer writer.Close()

	// 创建测试帧
	fmt.Println("创建测试视频...")
	for i := 0; i < 100; i++ { // 4秒的视频 (25fps * 4)
		frame := createTestFrame(i)
		if err := writer.WriteFrame(frame); err != nil {
			log.Fatalf("写入帧失败: %v", err)
		}

		if i%25 == 0 {
			fmt.Printf("进度: %d/100\n", i)
		}
	}

	fmt.Println("测试视频创建完成: test_video.mp4")
}

// createTestFrame 创建测试帧
func createTestFrame(frameNum int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 640, 480))

	// 创建简单的动画效果
	for y := 0; y < 480; y++ {
		for x := 0; x < 640; x++ {
			r := byte((x + frameNum) % 256)
			g := byte((y + frameNum) % 256)
			b := byte((x + y + frameNum) % 256)
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	return img
}
