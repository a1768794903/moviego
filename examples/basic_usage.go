package main

import (
	"fmt"
	"log"
	"time"

	"moviepy-go/pkg/ffmpeg"
	"moviepy-go/pkg/video"
)

func main() {
	// 创建进程管理器
	processMgr := ffmpeg.NewProcessManager()
	defer processMgr.Close()

	// 创建视频剪辑
	clip := video.NewVideoFileClip("./example.mp4", processMgr)

	// 打开视频文件
	if err := clip.Open(); err != nil {
		log.Fatalf("打开视频失败: %v", err)
	}
	defer clip.Close()

	// 打印视频信息
	fmt.Printf("视频信息:\n")
	fmt.Printf("  时长: %v\n", clip.Duration())
	fmt.Printf("  帧率: %.2f fps\n", clip.FPS())
	fmt.Printf("  尺寸: %dx%d\n", clip.Width(), clip.Height())
	fmt.Printf("  宽高比: %.2f\n", clip.AspectRatio())

	// 获取视频帧
	fmt.Printf("\n获取帧示例:\n")
	for i := 0; i < 3; i++ {
		t := time.Duration(i) * time.Second
		if t > clip.Duration() {
			break
		}

		frame, err := clip.GetFrame(t)
		if err != nil {
			log.Printf("获取帧失败 (t=%v): %v", t, err)
			continue
		}

		bounds := frame.Bounds()
		fmt.Printf("  时间 %v: 帧尺寸 %dx%d\n", t, bounds.Dx(), bounds.Dy())
	}

	// 创建子剪辑
	if clip.Duration() > 2*time.Second {
		fmt.Printf("\n创建子剪辑:\n")
		subclip, err := clip.Subclip(time.Second, 2*time.Second)
		if err != nil {
			log.Printf("创建子剪辑失败: %v", err)
		} else {
			fmt.Printf("  子剪辑时长: %v\n", subclip.Duration())
			subclip.Close()
		}
	}

	// 调整播放速度
	fmt.Printf("\n速度调整:\n")
	fastClip, err := clip.WithSpeed(2.0)
	if err != nil {
		log.Printf("速度调整失败: %v", err)
	} else {
		fmt.Printf("  2倍速剪辑时长: %v\n", fastClip.Duration())
		fastClip.Close()
	}

	// 调整音量
	fmt.Printf("\n音量调整:\n")
	volumeClip, err := clip.WithVolume(0.5)
	if err != nil {
		log.Printf("音量调整失败: %v", err)
	} else {
		fmt.Printf("  音量减半剪辑时长: %v\n", volumeClip.Duration())
		volumeClip.Close()
	}

	fmt.Println("\n示例完成!")
}
