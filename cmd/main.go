package main

import (
	"context"
	"fmt"
	"image"
	"log"
	"os"
	"time"

	"moviepy-go/pkg/ffmpeg"
	"moviepy-go/pkg/video"
)

func main() {
	// 创建进程管理器
	processMgr := ffmpeg.NewProcessManager()
	defer processMgr.Close()

	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("用法: moviepy-go <视频文件路径>")
		os.Exit(1)
	}

	filename := os.Args[1]

	fmt.Printf("开始处理视频文件: %s\n", filename)

	// 创建视频剪辑
	clip := video.NewVideoFileClip(filename, processMgr)

	// 打开视频文件
	fmt.Println("正在打开视频文件...")
	if err := clip.Open(); err != nil {
		log.Fatalf("打开视频失败: %v", err)
	}
	defer clip.Close()

	// 打印视频信息
	fmt.Printf("视频信息:\n")
	fmt.Printf("  文件名: %s\n", filename)
	fmt.Printf("  时长: %v\n", clip.Duration())
	fmt.Printf("  帧率: %.2f fps\n", clip.FPS())
	fmt.Printf("  尺寸: %dx%d\n", clip.Width(), clip.Height())
	fmt.Printf("  宽高比: %.2f\n", clip.AspectRatio())

	// 演示获取帧（带超时）
	fmt.Printf("\n获取帧示例:\n")
	for i := 0; i < 3; i++ {
		t := time.Duration(i) * time.Second
		if t > clip.Duration() {
			break
		}

		fmt.Printf("  正在获取第 %d 帧 (时间: %v)...\n", i+1, t)

		// 创建带超时的上下文
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		// 在 goroutine 中获取帧
		frameChan := make(chan interface{}, 1)
		errChan := make(chan error, 1)

		go func() {
			frame, err := clip.GetFrame(t)
			if err != nil {
				errChan <- err
				return
			}
			frameChan <- frame
		}()

		// 等待结果或超时
		select {
		case frame := <-frameChan:
			bounds := frame.(image.Image).Bounds()
			fmt.Printf("    成功: 帧尺寸 %dx%d\n", bounds.Dx(), bounds.Dy())
		case err := <-errChan:
			log.Printf("    失败: %v", err)
		case <-ctx.Done():
			fmt.Printf("    超时: 获取帧超时 (10秒)\n")
		}

		cancel()
	}

	// 演示子剪辑
	fmt.Printf("\n创建子剪辑:\n")
	if clip.Duration() > 2*time.Second {
		subclip, err := clip.Subclip(time.Second, 2*time.Second)
		if err != nil {
			log.Printf("创建子剪辑失败: %v", err)
		} else {
			fmt.Printf("  子剪辑时长: %v\n", subclip.Duration())
			subclip.Close()
		}
	}

	// 演示速度调整
	fmt.Printf("\n速度调整:\n")
	fastClip, err := clip.WithSpeed(2.0)
	if err != nil {
		log.Printf("速度调整失败: %v", err)
	} else {
		fmt.Printf("  2倍速剪辑时长: %v\n", fastClip.Duration())
		fastClip.Close()
	}

	// 演示音量调整
	fmt.Printf("\n音量调整:\n")
	volumeClip, err := clip.WithVolume(0.5)
	if err != nil {
		log.Printf("音量调整失败: %v", err)
	} else {
		fmt.Printf("  音量减半剪辑时长: %v\n", volumeClip.Duration())
		volumeClip.Close()
	}

	// 打印进程管理信息
	fmt.Printf("\n进程管理:\n")
	fmt.Printf("  当前管理的进程数: %d\n", processMgr.GetProcessCount())

	fmt.Println("\n演示完成!")
}
