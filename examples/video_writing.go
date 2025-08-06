package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"moviepy-go/pkg/core"
	"moviepy-go/pkg/ffmpeg"
	"moviepy-go/pkg/video"
)

func main() {
	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("用法: go run video_writing.go <输入视频文件路径> [输出视频文件路径]")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := "output.mp4"
	if len(os.Args) >= 3 {
		outputFile = os.Args[2]
	}

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

	// 打印视频信息
	fmt.Printf("输入视频信息:\n")
	fmt.Printf("  文件名: %s\n", inputFile)
	fmt.Printf("  时长: %v\n", clip.Duration())
	fmt.Printf("  帧率: %.2f fps\n", clip.FPS())
	fmt.Printf("  尺寸: %dx%d\n", clip.Width(), clip.Height())
	fmt.Printf("  宽高比: %.2f\n", clip.AspectRatio())

	// 设置写入选项
	options := &core.WriteOptions{
		Codec:   "libx264",
		Bitrate: "2000k",
		FPS:     clip.FPS(),
	}

	fmt.Printf("\n开始写入视频到: %s\n", outputFile)
	fmt.Printf("编码器: %s\n", options.Codec)
	fmt.Printf("比特率: %s\n", options.Bitrate)
	fmt.Printf("帧率: %.2f fps\n", options.FPS)

	// 写入视频文件
	startTime := time.Now()
	if err := clip.WriteToFile(outputFile, options); err != nil {
		log.Fatalf("写入视频失败: %v", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("\n视频写入完成!\n")
	fmt.Printf("输出文件: %s\n", outputFile)
	fmt.Printf("处理时间: %v\n", duration)

	// 演示子剪辑写入
	if clip.Duration() > 5*time.Second {
		fmt.Printf("\n创建子剪辑示例:\n")
		subclip, err := clip.Subclip(2*time.Second, 4*time.Second)
		if err != nil {
			log.Printf("创建子剪辑失败: %v", err)
		} else {
			subclipFile := "subclip.mp4"
			fmt.Printf("写入子剪辑到: %s\n", subclipFile)

			subclipOptions := &core.WriteOptions{
				Codec:   "libx264",
				Bitrate: "1500k",
				FPS:     subclip.FPS(),
			}

			if err := subclip.WriteToFile(subclipFile, subclipOptions); err != nil {
				log.Printf("写入子剪辑失败: %v", err)
			} else {
				fmt.Printf("子剪辑写入完成: %s\n", subclipFile)
			}
			subclip.Close()
		}
	}

	// 演示速度调整
	fmt.Printf("\n速度调整示例:\n")
	fastClip, err := clip.WithSpeed(2.0)
	if err != nil {
		log.Printf("速度调整失败: %v", err)
	} else {
		fastFile := "fast.mp4"
		fmt.Printf("写入2倍速视频到: %s\n", fastFile)

		fastOptions := &core.WriteOptions{
			Codec:   "libx264",
			Bitrate: "1000k",
			FPS:     fastClip.FPS(),
		}

		if err := fastClip.WriteToFile(fastFile, fastOptions); err != nil {
			log.Printf("写入快速视频失败: %v", err)
		} else {
			fmt.Printf("快速视频写入完成: %s\n", fastFile)
		}
		fastClip.Close()
	}

	fmt.Printf("\n所有示例完成!\n")
}
