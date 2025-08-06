package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"moviepy-go/pkg/audio"
	"moviepy-go/pkg/core"
	"moviepy-go/pkg/ffmpeg"
)

func main() {
	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("用法: go run audio_processing.go <音频文件路径>")
		os.Exit(1)
	}

	inputFile := os.Args[1]

	// 创建进程管理器
	processMgr := ffmpeg.NewProcessManager()
	defer processMgr.Close()

	// 创建音频剪辑
	clip := audio.NewAudioFileClip(inputFile, processMgr)

	// 打开音频文件
	if err := clip.Open(); err != nil {
		log.Printf("打开音频失败: %v", err)
		log.Println("创建测试音频...")

		// 创建一个简单的测试音频
		createTestAudio(processMgr)
		return
	}
	defer clip.Close()

	// 打印音频信息
	fmt.Printf("音频信息:\n")
	fmt.Printf("  文件名: %s\n", inputFile)
	fmt.Printf("  时长: %v\n", clip.Duration())
	fmt.Printf("  采样率: %d Hz\n", clip.SampleRate())
	fmt.Printf("  声道数: %d\n", clip.Channels())

	// 设置写入选项
	options := &core.WriteOptions{
		AudioCodec:   "aac",
		AudioBitrate: "128k",
	}

	// 写入音频文件
	fmt.Printf("\n开始写入音频到: output_audio.aac\n")
	if err := clip.WriteToFile("output_audio.aac", options); err != nil {
		log.Fatalf("写入音频失败: %v", err)
	}

	// 演示子剪辑
	fmt.Printf("\n创建子剪辑示例:\n")
	if clip.Duration() > 2*time.Second {
		subclip, err := clip.Subclip(time.Second, 2*time.Second)
		if err != nil {
			log.Printf("创建子剪辑失败: %v", err)
		} else {
			fmt.Printf("  子剪辑时长: %v\n", subclip.Duration())

			// 写入子剪辑
			subclipOptions := &core.WriteOptions{
				AudioCodec:   "mp3",
				AudioBitrate: "96k",
			}

			if err := subclip.WriteToFile("subclip_audio.mp3", subclipOptions); err != nil {
				log.Printf("写入子剪辑失败: %v", err)
			} else {
				fmt.Printf("  子剪辑写入完成: subclip_audio.mp3\n")
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
		fmt.Printf("  2倍速剪辑时长: %v\n", fastClip.Duration())

		fastOptions := &core.WriteOptions{
			AudioCodec:   "aac",
			AudioBitrate: "160k",
		}

		if err := fastClip.WriteToFile("fast_audio.aac", fastOptions); err != nil {
			log.Printf("写入快速音频失败: %v", err)
		} else {
			fmt.Printf("  快速音频写入完成: fast_audio.aac\n")
		}
		fastClip.Close()
	}

	// 演示音量调整
	fmt.Printf("\n音量调整示例:\n")
	volumeClip, err := clip.WithVolume(0.5)
	if err != nil {
		log.Printf("音量调整失败: %v", err)
	} else {
		fmt.Printf("  音量减半剪辑时长: %v\n", volumeClip.Duration())

		volumeOptions := &core.WriteOptions{
			AudioCodec:   "aac",
			AudioBitrate: "128k",
		}

		if err := volumeClip.WriteToFile("quiet_audio.aac", volumeOptions); err != nil {
			log.Printf("写入静音音频失败: %v", err)
		} else {
			fmt.Printf("  静音音频写入完成: quiet_audio.aac\n")
		}
		volumeClip.Close()
	}

	fmt.Printf("\n所有音频处理示例完成!\n")
}

// createTestAudio 创建一个简单的测试音频
func createTestAudio(processMgr *ffmpeg.ProcessManager) {
	// 创建音频写入器
	writerOptions := &ffmpeg.AudioWriterOptions{
		Codec:      "aac",
		Bitrate:    "128k",
		SampleRate: 44100,
		Channels:   2,
	}

	writer := ffmpeg.NewAudioWriter("test_audio.aac", writerOptions, processMgr)

	// 打开写入器
	if err := writer.Open(); err != nil {
		log.Fatalf("打开写入器失败: %v", err)
	}
	defer writer.Close()

	// 创建测试音频数据
	fmt.Println("创建测试音频...")
	sampleRate := 44100
	duration := 1                             // 1秒
	totalSamples := sampleRate * duration * 2 // 立体声

	// 每次写入较小的数据块
	chunkSize := 4410 // 0.1秒的数据

	for i := 0; i < totalSamples; i += chunkSize {
		end := i + chunkSize
		if end > totalSamples {
			end = totalSamples
		}

		samples := make([]float64, end-i)
		for j := range samples {
			// 创建简单的正弦波
			time := float64(i+j) / float64(sampleRate)
			frequency := 440.0                                 // A4音符
			sample := 0.1 * math.Sin(2*math.Pi*frequency*time) // 降低音量
			samples[j] = sample
		}

		if err := writer.WriteSamples(samples); err != nil {
			log.Fatalf("写入音频数据失败: %v", err)
		}

		progress := float64(i) / float64(totalSamples) * 100
		fmt.Printf("进度: %.1f%%\n", progress)
	}

	fmt.Println("测试音频创建完成: test_audio.aac")
}
