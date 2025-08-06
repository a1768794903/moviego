package core

import "errors"

// 错误定义
var (
	ErrNotImplemented      = errors.New("功能尚未实现")
	ErrInvalidTimeRange    = errors.New("无效的时间范围")
	ErrInvalidSpeedFactor  = errors.New("无效的速度因子")
	ErrInvalidVolumeFactor = errors.New("无效的音量因子")
	ErrFileNotFound        = errors.New("文件未找到")
	ErrInvalidFormat       = errors.New("无效的文件格式")
	ErrFFmpegError         = errors.New("FFmpeg 执行错误")
	ErrContextCancelled    = errors.New("上下文已取消")
	ErrResourceClosed      = errors.New("资源已关闭")
	ErrInvalidFrame        = errors.New("无效的帧数据")
	ErrInvalidAudioFrame   = errors.New("无效的音频帧数据")
	ErrUnsupportedCodec    = errors.New("不支持的编解码器")
	ErrMemoryLimit         = errors.New("内存使用超出限制")
	ErrProcessTerminated   = errors.New("进程被终止")
)
