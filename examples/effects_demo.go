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
	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("用法: go run effects_demo.go <输入视频文件路径>")
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

	// 演示各种特效
	demonstrateEffects(clip, processMgr)
}

// demonstrateEffects 演示各种特效
func demonstrateEffects(originalClip core.VideoClip, processMgr *ffmpeg.ProcessManager) {
	fmt.Printf("\n=== 特效演示 ===\n")

	// 1. 缩放特效
	fmt.Printf("\n1. 缩放特效 (640x480)\n")
	resizeEffect := effects.NewResizeEffect(640, 480)
	effectClip := video.NewEffectVideoClip(originalClip, processMgr)
	effectClip.AddEffect(resizeEffect)

	// 写入缩放后的视频
	options := &core.WriteOptions{
		Codec:   "libx264",
		Bitrate: "1500k",
		FPS:     originalClip.FPS(),
	}

	if err := effectClip.WriteToFile("resized_video.mp4", options); err != nil {
		log.Printf("写入缩放视频失败: %v", err)
	} else {
		fmt.Printf("  缩放视频写入完成: resized_video.mp4\n")
	}
	effectClip.Close()

	// 2. 旋转特效
	fmt.Printf("\n2. 旋转特效 (90度)\n")
	rotateEffect := effects.NewRotateEffect(90.0)
	effectClip = video.NewEffectVideoClip(originalClip, processMgr)
	effectClip.AddEffect(rotateEffect)

	if err := effectClip.WriteToFile("rotated_video.mp4", options); err != nil {
		log.Printf("写入旋转视频失败: %v", err)
	} else {
		fmt.Printf("  旋转视频写入完成: rotated_video.mp4\n")
	}
	effectClip.Close()

	// 3. 裁剪特效
	fmt.Printf("\n3. 裁剪特效 (中心区域)\n")
	// 计算中心裁剪区域
	width := originalClip.Width()
	height := originalClip.Height()
	cropWidth := width / 2
	cropHeight := height / 2
	cropX := (width - cropWidth) / 2
	cropY := (height - cropHeight) / 2

	cropEffect := effects.NewCropEffect(cropX, cropY, cropWidth, cropHeight)
	effectClip = video.NewEffectVideoClip(originalClip, processMgr)
	effectClip.AddEffect(cropEffect)

	if err := effectClip.WriteToFile("cropped_video.mp4", options); err != nil {
		log.Printf("写入裁剪视频失败: %v", err)
	} else {
		fmt.Printf("  裁剪视频写入完成: cropped_video.mp4\n")
	}
	effectClip.Close()

	// 4. 亮度调整特效
	fmt.Printf("\n4. 亮度调整特效 (1.5倍亮度)\n")
	brightnessEffect := effects.NewBrightnessEffect(1.5)
	effectClip = video.NewEffectVideoClip(originalClip, processMgr)
	effectClip.AddEffect(brightnessEffect)

	if err := effectClip.WriteToFile("brightened_video.mp4", options); err != nil {
		log.Printf("写入亮度调整视频失败: %v", err)
	} else {
		fmt.Printf("  亮度调整视频写入完成: brightened_video.mp4\n")
	}
	effectClip.Close()

	// 5. 对比度调整特效
	fmt.Printf("\n5. 对比度调整特效 (1.3倍对比度)\n")
	contrastEffect := effects.NewContrastEffect(1.3)
	effectClip = video.NewEffectVideoClip(originalClip, processMgr)
	effectClip.AddEffect(contrastEffect)

	if err := effectClip.WriteToFile("contrast_video.mp4", options); err != nil {
		log.Printf("写入对比度调整视频失败: %v", err)
	} else {
		fmt.Printf("  对比度调整视频写入完成: contrast_video.mp4\n")
	}
	effectClip.Close()

	// 6. 组合特效
	fmt.Printf("\n6. 组合特效 (缩放 + 旋转 + 亮度调整)\n")
	effectClip = video.NewEffectVideoClip(originalClip, processMgr)
	effectClip.AddEffect(effects.NewResizeEffect(480, 360))
	effectClip.AddEffect(effects.NewRotateEffect(45.0))
	effectClip.AddEffect(effects.NewBrightnessEffect(1.2))

	if err := effectClip.WriteToFile("combined_effects_video.mp4", options); err != nil {
		log.Printf("写入组合特效视频失败: %v", err)
	} else {
		fmt.Printf("  组合特效视频写入完成: combined_effects_video.mp4\n")
	}
	effectClip.Close()

	// 7. 子剪辑特效
	fmt.Printf("\n7. 子剪辑特效 (前5秒 + 旋转)\n")
	if originalClip.Duration() > 5*time.Second {
		subclip, err := originalClip.Subclip(0, 5*time.Second)
		if err != nil {
			log.Printf("创建子剪辑失败: %v", err)
		} else {
			videoSubclip, ok := subclip.(core.VideoClip)
			if ok {
				effectClip = video.NewEffectVideoClip(videoSubclip, processMgr)
				effectClip.AddEffect(effects.NewRotateEffect(180.0))

				subclipOptions := &core.WriteOptions{
					Codec:   "libx264",
					Bitrate: "1000k",
					FPS:     videoSubclip.FPS(),
				}

				if err := effectClip.WriteToFile("subclip_rotated_video.mp4", subclipOptions); err != nil {
					log.Printf("写入子剪辑特效视频失败: %v", err)
				} else {
					fmt.Printf("  子剪辑特效视频写入完成: subclip_rotated_video.mp4\n")
				}
				effectClip.Close()
			}
		}
	}

	fmt.Printf("\n=== 特效演示完成 ===\n")
	fmt.Printf("生成的文件:\n")
	fmt.Printf("  - resized_video.mp4 (缩放)\n")
	fmt.Printf("  - rotated_video.mp4 (旋转)\n")
	fmt.Printf("  - cropped_video.mp4 (裁剪)\n")
	fmt.Printf("  - brightened_video.mp4 (亮度调整)\n")
	fmt.Printf("  - contrast_video.mp4 (对比度调整)\n")
	fmt.Printf("  - combined_effects_video.mp4 (组合特效)\n")
	if originalClip.Duration() > 5*time.Second {
		fmt.Printf("  - subclip_rotated_video.mp4 (子剪辑特效)\n")
	}
}
