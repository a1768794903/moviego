# MoviePy Go - Makefile

.PHONY: all build test run clean deps fmt lint cross-build examples video-writing audio-processing effects-demo compositing-demo run-basic

# 默认目标
all: build

# 构建主程序
build:
	@echo "构建 MoviePy Go..."
	go build -o bin/moviepy-go cmd/main.go

# 构建所有示例
examples: build
	@echo "构建示例程序..."
	go build -o bin/basic_usage examples/basic_usage.go
	go build -o bin/video_writing examples/video_writing.go
	go build -o bin/simple_writer examples/simple_writer.go
	go build -o bin/audio_processing examples/audio_processing.go
	go build -o bin/effects_demo examples/effects_demo.go
	go build -o bin/compositing_demo examples/compositing_demo.go

# 运行视频写入示例
video-writing: examples
	@echo "运行视频写入示例..."
	@if [ -z "$(INPUT)" ]; then \
		echo "请指定输入视频文件: make video-writing INPUT=video.mp4"; \
		exit 1; \
	fi
	@if [ -z "$(OUTPUT)" ]; then \
		./bin/video_writing $(INPUT); \
	else \
		./bin/video_writing $(INPUT) $(OUTPUT); \
	fi

# 运行基本示例
run-basic: examples
	@echo "运行基本示例..."
	@if [ -z "$(INPUT)" ]; then \
		echo "请指定输入视频文件: make run-basic INPUT=video.mp4"; \
		exit 1; \
	fi
	./bin/basic_usage $(INPUT)

# 运行主程序
run: build
	@echo "运行主程序..."
	@if [ -z "$(INPUT)" ]; then \
		echo "请指定输入视频文件: make run INPUT=video.mp4"; \
		exit 1; \
	fi
	./bin/moviepy-go $(INPUT)

# 测试
test:
	@echo "运行测试..."
	go test -v ./...

# 运行简单测试
test-simple: examples
	@echo "运行简单测试..."
	./bin/simple_writer

# 清理
clean:
	@echo "清理构建文件..."
	rm -rf bin/
	go clean

# 依赖管理
deps:
	@echo "下载依赖..."
	go mod download
	go mod tidy

# 代码格式化
fmt:
	@echo "格式化代码..."
	go fmt ./...

# 代码检查
lint:
	@echo "检查代码..."
	golangci-lint run

# 交叉编译
cross-build:
	@echo "交叉编译..."
	GOOS=linux GOARCH=amd64 go build -o bin/moviepy-go-linux-amd64 cmd/main.go
	GOOS=windows GOARCH=amd64 go build -o bin/moviepy-go-windows-amd64.exe cmd/main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/moviepy-go-darwin-amd64 cmd/main.go

# 安装依赖
install-deps:
	@echo "安装开发依赖..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 帮助
help:
	@echo "可用的目标:"
	@echo "  build        - 构建主程序"
	@echo "  examples     - 构建所有示例"
	@echo "  video-writing - 运行视频写入示例 (需要 INPUT=video.mp4)"
	@echo "  audio-processing - 运行音频处理示例 (需要 INPUT=audio.mp3)"
	@echo "  effects-demo  - 运行特效演示示例 (需要 INPUT=video.mp4)"
	@echo "  compositing-demo - 运行合成演示示例 (需要 INPUT=video.mp4)"
	@echo "  run-basic    - 运行基本示例 (需要 INPUT=video.mp4)"
	@echo "  run          - 运行主程序 (需要 INPUT=video.mp4)"
	@echo "  test         - 运行测试"
	@echo "  clean        - 清理构建文件"
	@echo "  deps         - 下载依赖"
	@echo "  fmt          - 格式化代码"
	@echo "  lint         - 检查代码"
	@echo "  cross-build  - 交叉编译"
	@echo "  install-deps - 安装开发依赖"
	@echo "  help         - 显示此帮助"

# 运行音频处理示例
audio-processing: examples
	@echo "运行音频处理示例..."
	@if [ -z "$(INPUT)" ]; then \
		echo "请指定输入音频文件: make audio-processing INPUT=audio.mp3"; \
		exit 1; \
	fi
	./bin/audio_processing $(INPUT)

# 运行特效演示示例
effects-demo: examples
	@echo "运行特效演示示例..."
	@if [ -z "$(INPUT)" ]; then \
		echo "请指定输入视频文件: make effects-demo INPUT=video.mp4"; \
		exit 1; \
	fi
	./bin/effects_demo $(INPUT)

# 运行合成演示示例
compositing-demo: examples
	@echo "运行合成演示示例..."
	@if [ -z "$(INPUT)" ]; then \
		echo "请指定输入视频文件: make compositing-demo INPUT=video.mp4"; \
		exit 1; \
	fi
	./bin/compositing_demo $(INPUT) 