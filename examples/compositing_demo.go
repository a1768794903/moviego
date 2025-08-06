package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"moviepy-go/pkg/compositing"
	"moviepy-go/pkg/core"
	"moviepy-go/pkg/ffmpeg"
	"moviepy-go/pkg/video"
)

func main() {
	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("用法: go run compositing_demo.go <输入视频文件路径>")
		os.Exit(1)
	}

	inputFile := os.Args[1]

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

	// 打印原始视频信息
	fmt.Printf("原始视频信息:\n")
	fmt.Printf("  文件名: %s\n", inputFile)
	fmt.Printf("  时长: %v\n", clip.Duration())
	fmt.Printf("  帧率: %.2f fps\n", clip.FPS())
	fmt.Printf("  尺寸: %dx%d\n", clip.Width(), clip.Height())
	fmt.Printf("  宽高比: %.2f\n", clip.AspectRatio())

	// 演示各种合成效果
	demonstrateCompositing(clip, processMgr)
}

// demonstrateCompositing 演示各种合成效果
func demonstrateCompositing(originalClip core.VideoClip, processMgr *ffmpeg.ProcessManager) {
	fmt.Printf("\n=== 合成演示 ===\n")

	// 创建剪辑数组函数
	createClips := func() []core.VideoClip {
		// 现在可以安全地重复使用同一个剪辑，因为CompositeVideoClip不会关闭它们
		return []core.VideoClip{originalClip, originalClip}
	}

	// 1. 叠加合成
	fmt.Printf("\n1. 叠加合成 (Overlay)\n")
	clips := createClips()
	positions := []*compositing.Position{
		compositing.NewPosition(0, 0),
		compositing.NewPosition(50, 50),
	}

	compositeClip := compositing.NewCompositeVideoClip(clips, positions, compositing.Overlay, processMgr)

	options := &core.WriteOptions{
		Codec:   "libx264",
		Bitrate: "2000k",
		FPS:     originalClip.FPS(),
	}

	if err := compositeClip.WriteToFile("overlay_composite.mp4", options); err != nil {
		log.Printf("写入叠加合成视频失败: %v", err)
	} else {
		fmt.Printf("  叠加合成视频写入完成: overlay_composite.mp4\n")
	}
	compositeClip.Close()

	// 2. 相加合成
	fmt.Printf("\n2. 相加合成 (Add)\n")
	clips = createClips()
	compositeClip = compositing.NewCompositeVideoClip(clips, positions, compositing.Add, processMgr)

	if err := compositeClip.WriteToFile("add_composite.mp4", options); err != nil {
		log.Printf("写入相加合成视频失败: %v", err)
	} else {
		fmt.Printf("  相加合成视频写入完成: add_composite.mp4\n")
	}
	compositeClip.Close()

	// 3. 相乘合成
	fmt.Printf("\n3. 相乘合成 (Multiply)\n")
	clips = createClips()
	compositeClip = compositing.NewCompositeVideoClip(clips, positions, compositing.Multiply, processMgr)

	if err := compositeClip.WriteToFile("multiply_composite.mp4", options); err != nil {
		log.Printf("写入相乘合成视频失败: %v", err)
	} else {
		fmt.Printf("  相乘合成视频写入完成: multiply_composite.mp4\n")
	}
	compositeClip.Close()

	// 4. 屏幕合成
	fmt.Printf("\n4. 屏幕合成 (Screen)\n")
	clips = createClips()
	compositeClip = compositing.NewCompositeVideoClip(clips, positions, compositing.Screen, processMgr)

	if err := compositeClip.WriteToFile("screen_composite.mp4", options); err != nil {
		log.Printf("写入屏幕合成视频失败: %v", err)
	} else {
		fmt.Printf("  屏幕合成视频写入完成: screen_composite.mp4\n")
	}
	compositeClip.Close()

	// 5. 变暗合成
	fmt.Printf("\n5. 变暗合成 (Darken)\n")
	clips = createClips()
	compositeClip = compositing.NewCompositeVideoClip(clips, positions, compositing.Darken, processMgr)

	if err := compositeClip.WriteToFile("darken_composite.mp4", options); err != nil {
		log.Printf("写入变暗合成视频失败: %v", err)
	} else {
		fmt.Printf("  变暗合成视频写入完成: darken_composite.mp4\n")
	}
	compositeClip.Close()

	// 6. 变亮合成
	fmt.Printf("\n6. 变亮合成 (Lighten)\n")
	clips = createClips()
	compositeClip = compositing.NewCompositeVideoClip(clips, positions, compositing.Lighten, processMgr)

	if err := compositeClip.WriteToFile("lighten_composite.mp4", options); err != nil {
		log.Printf("写入变亮合成视频失败: %v", err)
	} else {
		fmt.Printf("  变亮合成视频写入完成: lighten_composite.mp4\n")
	}
	compositeClip.Close()

	// 7. 多剪辑合成
	fmt.Printf("\n7. 多剪辑合成\n")
	if originalClip.Duration() > 3*time.Second {
		// 创建子剪辑
		subclip1, err := originalClip.Subclip(0, 2*time.Second)
		if err != nil {
			log.Printf("创建子剪辑1失败: %v", err)
		} else {
			subclip2, err := originalClip.Subclip(1*time.Second, 3*time.Second)
			if err != nil {
				log.Printf("创建子剪辑2失败: %v", err)
			} else {
				videoSubclip1, ok1 := subclip1.(core.VideoClip)
				videoSubclip2, ok2 := subclip2.(core.VideoClip)

				if ok1 && ok2 {
					multiClips := []core.VideoClip{originalClip, videoSubclip1, videoSubclip2}
					multiPositions := []*compositing.Position{
						compositing.NewPosition(0, 0),
						compositing.NewPosition(100, 100),
						compositing.NewPosition(200, 200),
					}

					multiComposite := compositing.NewCompositeVideoClip(multiClips, multiPositions, compositing.Overlay, processMgr)

					if err := multiComposite.WriteToFile("multi_composite.mp4", options); err != nil {
						log.Printf("写入多剪辑合成视频失败: %v", err)
					} else {
						fmt.Printf("  多剪辑合成视频写入完成: multi_composite.mp4\n")
					}
					multiComposite.Close()
				}
			}
		}
	}

	// 8. 居中合成
	fmt.Printf("\n8. 居中合成\n")
	centeredClips := createClips()
	centeredPositions := []*compositing.Position{
		compositing.NewPosition(0, 0),
		compositing.NewCenteredPosition(),
	}

	centeredComposite := compositing.NewCompositeVideoClip(centeredClips, centeredPositions, compositing.Overlay, processMgr)

	if err := centeredComposite.WriteToFile("centered_composite.mp4", options); err != nil {
		log.Printf("写入居中合成视频失败: %v", err)
	} else {
		fmt.Printf("  居中合成视频写入完成: centered_composite.mp4\n")
	}
	centeredComposite.Close()

	fmt.Printf("\n=== 合成演示完成 ===\n")
	fmt.Printf("生成的文件:\n")
	fmt.Printf("  - overlay_composite.mp4 (叠加合成)\n")
	fmt.Printf("  - add_composite.mp4 (相加合成)\n")
	fmt.Printf("  - multiply_composite.mp4 (相乘合成)\n")
	fmt.Printf("  - screen_composite.mp4 (屏幕合成)\n")
	fmt.Printf("  - darken_composite.mp4 (变暗合成)\n")
	fmt.Printf("  - lighten_composite.mp4 (变亮合成)\n")
	fmt.Printf("  - centered_composite.mp4 (居中合成)\n")
	if originalClip.Duration() > 3*time.Second {
		fmt.Printf("  - multi_composite.mp4 (多剪辑合成)\n")
	}
}
