# MovieGo

MovieGo 是用 Golang 重写的 MoviePy 项目，专注于解决原 Python 版本中的僵尸进程问题，并提供高性能的视频处理能力。

## 特性

- 🎬 **视频处理**: 支持视频读取、写入、剪辑、速度调整
- 🔊 **音频处理**: 支持音频读取、写入、剪辑、音量调整
- ✨ **特效处理**: 支持缩放、旋转、裁剪、亮度、对比度调整
- 🎭 **视频合成**: 支持多种合成模式（叠加、相加、相乘、屏幕、变暗、变亮）
- ⚡ **高性能**: 基于 Go 语言，提供更好的并发性能
- 🛡️ **进程管理**: 主动管理 FFmpeg 进程，解决僵尸进程问题
- 🔧 **类型安全**: 编译时类型检查，减少运行时错误
- 📦 **单一可执行文件**: 无需 Python 环境，部署简单
- 🔧 **模块化设计**: 清晰的接口和组件分离
- 📦 **易于使用**: 简洁的 API 设计

## 架构

```
moviego/
├── pkg/
│   ├── core/           # 核心接口和基础实现
│   ├── ffmpeg/         # FFmpeg 集成和进程管理
│   ├── video/          # 视频处理模块
│   └── audio/          # 音频处理模块
├── cmd/                # 主程序入口
├── examples/           # 示例代码
└── tests/              # 测试文件
```

## 安装

### 前置要求

- Go 1.21 或更高版本
- FFmpeg (需要安装并添加到 PATH)

### 安装步骤

1. 克隆仓库：
```bash
git clone https://github.com/a1768794903/moviego.git
cd moviego
```

2. 安装依赖：
```bash
make deps
```

3. 构建项目：
```bash
make build
```

## 使用方法

### 基本用法

```go
package main

import (
    "log"
    "moviego/pkg/ffmpeg"
    "moviego/pkg/video"
)

func main() {
    // 创建进程管理器
    processMgr := ffmpeg.NewProcessManager()
    defer processMgr.Close()

    // 创建视频剪辑
    clip := video.NewVideoFileClip("input.mp4", processMgr)
    
    // 打开视频
    if err := clip.Open(); err != nil {
        log.Fatal(err)
    }
    defer clip.Close()

    // 获取视频信息
    fmt.Printf("时长: %v\n", clip.Duration())
    fmt.Printf("尺寸: %dx%d\n", clip.Width(), clip.Height())
    fmt.Printf("帧率: %.2f fps\n", clip.FPS())
}
```

### 视频写入

```go
// 设置写入选项
options := &core.WriteOptions{
    Codec:   "libx264",
    Bitrate: "2000k",
    FPS:     25.0,
}

// 写入视频文件
if err := clip.WriteToFile("output.mp4", options); err != nil {
    log.Fatal(err)
}
```

### 视频剪辑操作

```go
// 创建子剪辑
subclip, err := clip.Subclip(2*time.Second, 5*time.Second)
if err != nil {
    log.Fatal(err)
}
defer subclip.Close()

// 调整播放速度
fastClip, err := clip.WithSpeed(2.0)
if err != nil {
    log.Fatal(err)
}
defer fastClip.Close()

// 调整音量
volumeClip, err := clip.WithVolume(0.5)
if err != nil {
    log.Fatal(err)
}
defer volumeClip.Close()
```

## 示例

### 运行基本示例

```bash
make run-basic INPUT=video.mp4
```

### 运行视频写入示例

```bash
make video-writing INPUT=video.mp4 OUTPUT=output.mp4
```

### 运行主程序

```bash
make run INPUT=video.mp4
```

### 运行特效演示示例

```bash
make effects-demo INPUT=video.mp4
```

### 运行合成演示示例

```bash
make compositing-demo INPUT=video.mp4
```

## 与 Python MoviePy 的对比

### 解决的问题

**原 Python MoviePy 问题**:
- 循环导入导致的僵尸进程
- 依赖垃圾回收器清理资源
- 进程管理不够主动

**Go 版本解决方案**:
- 主动进程管理 (`ProcessManager`)
- 使用 `context.Context` 进行取消控制
- 显式资源清理
- 无循环导入问题

### 性能优势

- **并发处理**: Go 的 goroutine 提供更好的并发性能
- **内存管理**: Go 的垃圾回收器更高效
- **进程控制**: 更精确的 FFmpeg 进程管理
- **类型安全**: 编译时类型检查

## 开发计划

- [x] 视频读取功能
- [x] 视频写入功能
- [x] 音频处理完善
- [x] 特效处理模块
- [x] 合成功能
- [ ] 更多格式支持
- [ ] 性能优化
- [ ] 测试覆盖

## 贡献

欢迎贡献代码！请查看 [CONTRIBUTING.md](CONTRIBUTING.md) 了解详情。

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 致谢

- 原 [MoviePy](https://github.com/Zulko/moviepy) 项目
- [FFmpeg](https://ffmpeg.org/) 团队
- Go 语言社区
