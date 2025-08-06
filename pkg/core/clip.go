package core

import (
	"context"
	"image"
	"time"
)

// Clip 是视频和音频剪辑的基类接口
type Clip interface {
	// 基础属性
	Duration() time.Duration
	Start() time.Duration
	End() time.Duration
	FPS() float64

	// 帧获取
	GetFrame(t time.Duration) (image.Image, error)
	GetAudioFrame(t time.Duration) ([]float64, error)

	// 变换操作
	Subclip(start, end time.Duration) (Clip, error)
	WithSpeed(factor float64) (Clip, error)
	WithVolume(factor float64) (Clip, error)

	// 合成操作
	WithAudio(audio AudioClip) (Clip, error)
	WithoutAudio() (Clip, error)

	// 写入操作
	WriteToFile(filename string, options *WriteOptions) error

	// 资源管理
	Close() error

	// 上下文管理器支持
	WithContext(ctx context.Context) Clip
}

// WriteOptions 写入选项
type WriteOptions struct {
	Codec        string
	Bitrate      string
	FPS          float64
	AudioCodec   string
	AudioBitrate string
}

// BaseClip 提供 Clip 接口的基础实现
type BaseClip struct {
	start    time.Duration
	end      time.Duration
	duration time.Duration
	fps      float64
	ctx      context.Context
}

// NewBaseClip 创建新的基础剪辑
func NewBaseClip(start, end, duration time.Duration, fps float64) *BaseClip {
	return &BaseClip{
		start:    start,
		end:      end,
		duration: duration,
		fps:      fps,
		ctx:      context.Background(),
	}
}

// Duration 获取剪辑时长
func (bc *BaseClip) Duration() time.Duration {
	return bc.duration
}

// Start 获取开始时间
func (bc *BaseClip) Start() time.Duration {
	return bc.start
}

// End 获取结束时间
func (bc *BaseClip) End() time.Duration {
	return bc.end
}

// FPS 获取帧率
func (bc *BaseClip) FPS() float64 {
	return bc.fps
}

// GetFrame 获取帧（基础实现返回错误）
func (bc *BaseClip) GetFrame(t time.Duration) (image.Image, error) {
	return nil, ErrNotImplemented
}

// GetAudioFrame 获取音频帧（基础实现返回错误）
func (bc *BaseClip) GetAudioFrame(t time.Duration) ([]float64, error) {
	return nil, ErrNotImplemented
}

// Subclip 创建子剪辑（基础实现返回错误）
func (bc *BaseClip) Subclip(start, end time.Duration) (Clip, error) {
	return nil, ErrNotImplemented
}

// WithSpeed 调整速度（基础实现返回错误）
func (bc *BaseClip) WithSpeed(factor float64) (Clip, error) {
	return nil, ErrNotImplemented
}

// WithVolume 调整音量（基础实现返回错误）
func (bc *BaseClip) WithVolume(factor float64) (Clip, error) {
	return nil, ErrNotImplemented
}

// WithAudio 添加音频（基础实现返回错误）
func (bc *BaseClip) WithAudio(audio AudioClip) (Clip, error) {
	return nil, ErrNotImplemented
}

// WithoutAudio 移除音频（基础实现返回错误）
func (bc *BaseClip) WithoutAudio() (Clip, error) {
	return nil, ErrNotImplemented
}

// WriteToFile 写入文件（基础实现返回错误）
func (bc *BaseClip) WriteToFile(filename string, options *WriteOptions) error {
	return ErrNotImplemented
}

// Close 关闭剪辑（基础实现）
func (bc *BaseClip) Close() error {
	return nil
}

// WithContext 设置上下文（基础实现）
func (bc *BaseClip) WithContext(ctx context.Context) Clip {
	bc.ctx = ctx
	return bc
}
