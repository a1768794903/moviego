package effects

import (
	"fmt"
	"image"

	"moviepy-go/pkg/core"
)

// EffectChain 特效链，可以组合多个特效
type EffectChain struct {
	effects []VideoEffect
}

// NewEffectChain 创建新的特效链
func NewEffectChain() *EffectChain {
	return &EffectChain{
		effects: make([]VideoEffect, 0),
	}
}

// AddEffect 添加特效到链中
func (ec *EffectChain) AddEffect(effect VideoEffect) {
	ec.effects = append(ec.effects, effect)
}

// ApplyToFrame 应用特效链到帧
func (ec *EffectChain) ApplyToFrame(frame image.Image) (image.Image, error) {
	result := frame

	for i, effect := range ec.effects {
		var err error
		result, err = effect.ApplyToFrame(result)
		if err != nil {
			return nil, fmt.Errorf("应用特效 %d (%s) 失败: %w", i, effect.GetName(), err)
		}
	}

	return result, nil
}

// GetEffects 获取所有特效
func (ec *EffectChain) GetEffects() []VideoEffect {
	return ec.effects
}

// Clear 清空特效链
func (ec *EffectChain) Clear() {
	ec.effects = make([]VideoEffect, 0)
}

// GetName 获取特效链名称
func (ec *EffectChain) GetName() string {
	return "effect_chain"
}

// Apply 应用特效链到剪辑
func (ec *EffectChain) Apply(clip core.Clip) (core.Clip, error) {
	// 这里应该返回一个新的剪辑，应用了特效链
	// 简化实现，直接返回原剪辑
	return clip, nil
}

// CompositeEffect 复合特效，可以组合多个特效链
type CompositeEffect struct {
	TransformEffect
	chains []*EffectChain
}

// NewCompositeEffect 创建复合特效
func NewCompositeEffect() *CompositeEffect {
	return &CompositeEffect{
		TransformEffect: TransformEffect{name: "composite"},
		chains:          make([]*EffectChain, 0),
	}
}

// AddChain 添加特效链
func (ce *CompositeEffect) AddChain(chain *EffectChain) {
	ce.chains = append(ce.chains, chain)
}

// ApplyToFrame 应用复合特效到帧
func (ce *CompositeEffect) ApplyToFrame(frame image.Image) (image.Image, error) {
	result := frame

	for i, chain := range ce.chains {
		var err error
		result, err = chain.ApplyToFrame(result)
		if err != nil {
			return nil, fmt.Errorf("应用特效链 %d 失败: %w", i, err)
		}
	}

	return result, nil
}

// Apply 应用复合特效到剪辑
func (ce *CompositeEffect) Apply(clip core.Clip) (core.Clip, error) {
	// 这里应该返回一个新的剪辑，应用了复合特效
	// 简化实现，直接返回原剪辑
	return clip, nil
}

// EffectBuilder 特效构建器，提供流畅的API
type EffectBuilder struct {
	chain *EffectChain
}

// NewEffectBuilder 创建特效构建器
func NewEffectBuilder() *EffectBuilder {
	return &EffectBuilder{
		chain: NewEffectChain(),
	}
}

// Resize 添加缩放特效
func (eb *EffectBuilder) Resize(width, height int) *EffectBuilder {
	eb.chain.AddEffect(NewResizeEffect(width, height))
	return eb
}

// Rotate 添加旋转特效
func (eb *EffectBuilder) Rotate(angle float64) *EffectBuilder {
	eb.chain.AddEffect(NewRotateEffect(angle))
	return eb
}

// Crop 添加裁剪特效
func (eb *EffectBuilder) Crop(x, y, width, height int) *EffectBuilder {
	eb.chain.AddEffect(NewCropEffect(x, y, width, height))
	return eb
}

// Brightness 添加亮度调整特效
func (eb *EffectBuilder) Brightness(factor float64) *EffectBuilder {
	eb.chain.AddEffect(NewBrightnessEffect(factor))
	return eb
}

// Contrast 添加对比度调整特效
func (eb *EffectBuilder) Contrast(factor float64) *EffectBuilder {
	eb.chain.AddEffect(NewContrastEffect(factor))
	return eb
}

// Blur 添加模糊特效
func (eb *EffectBuilder) Blur(radius int) *EffectBuilder {
	eb.chain.AddEffect(NewBlurEffect(radius))
	return eb
}

// Sharpen 添加锐化特效
func (eb *EffectBuilder) Sharpen(strength float64) *EffectBuilder {
	eb.chain.AddEffect(NewSharpenEffect(strength))
	return eb
}

// Saturation 添加饱和度调整特效
func (eb *EffectBuilder) Saturation(factor float64) *EffectBuilder {
	eb.chain.AddEffect(NewSaturationEffect(factor))
	return eb
}

// Noise 添加噪点特效
func (eb *EffectBuilder) Noise(intensity float64) *EffectBuilder {
	eb.chain.AddEffect(NewNoiseEffect(intensity))
	return eb
}

// Sepia 添加棕褐色特效
func (eb *EffectBuilder) Sepia(strength float64) *EffectBuilder {
	eb.chain.AddEffect(NewSepiaEffect(strength))
	return eb
}

// Vignette 添加暗角特效
func (eb *EffectBuilder) Vignette(strength, radius float64) *EffectBuilder {
	eb.chain.AddEffect(NewVignetteEffect(strength, radius))
	return eb
}

// Build 构建特效链
func (eb *EffectBuilder) Build() *EffectChain {
	return eb.chain
}

// EffectPreset 特效预设
type EffectPreset struct {
	name    string
	builder *EffectBuilder
}

// NewEffectPreset 创建特效预设
func NewEffectPreset(name string) *EffectPreset {
	return &EffectPreset{
		name:    name,
		builder: NewEffectBuilder(),
	}
}

// Vintage 复古预设
func Vintage() *EffectChain {
	return NewEffectBuilder().
		Sepia(0.8).
		Vignette(0.3, 0.8).
		Noise(0.1).
		Build()
}

// Cinematic 电影预设
func Cinematic() *EffectChain {
	return NewEffectBuilder().
		Contrast(1.2).
		Saturation(0.8).
		Vignette(0.4, 0.7).
		Build()
}

// Warm 暖色调预设
func Warm() *EffectChain {
	return NewEffectBuilder().
		Brightness(1.1).
		Saturation(1.2).
		Build()
}

// Cool 冷色调预设
func Cool() *EffectChain {
	return NewEffectBuilder().
		Brightness(0.9).
		Saturation(0.8).
		Build()
}

// Dramatic 戏剧性预设
func Dramatic() *EffectChain {
	return NewEffectBuilder().
		Contrast(1.5).
		Brightness(0.8).
		Vignette(0.6, 0.6).
		Build()
}
