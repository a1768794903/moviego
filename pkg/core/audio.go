package core

import (
	"time"
)

// AudioClip 音频剪辑接口
type AudioClip interface {
	Clip

	// 音频特有属性
	Channels() int
	SampleRate() int

	// 音频特有操作
	WithChannels(channels int) (AudioClip, error)
	WithSampleRate(sampleRate int) (AudioClip, error)
	Concatenate(other AudioClip) (AudioClip, error)
	Mix(other AudioClip) (AudioClip, error)
}

// BaseAudioClip 音频剪辑基础实现
type BaseAudioClip struct {
	*BaseClip
	channels   int
	sampleRate int
}

// NewBaseAudioClip 创建新的音频剪辑
func NewBaseAudioClip(start, end, duration time.Duration, fps float64, channels, sampleRate int) *BaseAudioClip {
	return &BaseAudioClip{
		BaseClip:   NewBaseClip(start, end, duration, fps),
		channels:   channels,
		sampleRate: sampleRate,
	}
}

// Channels 返回声道数
func (ac *BaseAudioClip) Channels() int {
	return ac.channels
}

// SampleRate 返回采样率
func (ac *BaseAudioClip) SampleRate() int {
	return ac.sampleRate
}

// WithChannels 设置声道数
func (ac *BaseAudioClip) WithChannels(channels int) (AudioClip, error) {
	if channels <= 0 {
		return nil, ErrInvalidFormat
	}

	// 这里应该返回一个新的音频剪辑，但基础实现返回错误
	return nil, ErrNotImplemented
}

// WithSampleRate 设置采样率
func (ac *BaseAudioClip) WithSampleRate(sampleRate int) (AudioClip, error) {
	if sampleRate <= 0 {
		return nil, ErrInvalidFormat
	}

	// 这里应该返回一个新的音频剪辑，但基础实现返回错误
	return nil, ErrNotImplemented
}

// Concatenate 连接音频剪辑
func (ac *BaseAudioClip) Concatenate(other AudioClip) (AudioClip, error) {
	// 这里应该返回一个新的音频剪辑，但基础实现返回错误
	return nil, ErrNotImplemented
}

// Mix 混合音频剪辑
func (ac *BaseAudioClip) Mix(other AudioClip) (AudioClip, error) {
	// 这里应该返回一个新的音频剪辑，但基础实现返回错误
	return nil, ErrNotImplemented
}

// GetAudioFrame 获取音频帧（基础实现）
func (ac *BaseAudioClip) GetAudioFrame(t time.Duration) ([]float64, error) {
	// 基础实现返回静音
	frameSize := int(float64(ac.sampleRate) * float64(time.Second) / ac.fps)
	samples := make([]float64, frameSize*ac.channels)
	return samples, nil
}
