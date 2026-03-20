package effects

import (
	"fmt"
	"strings"
	"time"
)

// Target specifies what streams an effect applies to.
type Target int

const (
	// TargetVideo means the effect applies to the video stream.
	TargetVideo Target = iota
	// TargetAudio means the effect applies to the audio stream.
	TargetAudio
	// TargetBoth means the effect applies to both video and audio.
	TargetBoth
)

// Effect is the interface for all composable effects.
// Effects describe a transformation that the timeline compiler translates
// into FFmpeg filter graph nodes.
type Effect interface {
	// Name returns a human-readable name for the effect (e.g. "fade_in", "volume").
	Name() string

	// Target returns what streams this effect applies to.
	Target() Target

	// FilterName returns the FFmpeg filter name (e.g. "fade", "volume", "setpts").
	FilterName() string

	// FilterParams returns the FFmpeg filter parameters as key-value pairs.
	FilterParams() map[string]string
}

// --- Video Effects ---

// FadeInEffect fades the video from black (or transparent) at the start.
type FadeInEffect struct {
	duration time.Duration
}

// FadeIn creates a video fade-in effect.
func FadeIn(d time.Duration) *FadeInEffect {
	return &FadeInEffect{duration: d}
}

func (e *FadeInEffect) Name() string       { return "fade_in" }
func (e *FadeInEffect) Target() Target     { return TargetVideo }
func (e *FadeInEffect) FilterName() string { return "fade" }
func (e *FadeInEffect) FilterParams() map[string]string {
	return map[string]string{
		"t":  "in",
		"st": "0",
		"d":  formatSeconds(e.duration),
	}
}

// Dur returns the fade duration.
func (e *FadeInEffect) Dur() time.Duration { return e.duration }

// FadeOutEffect fades the video to black (or transparent) at a given time.
type FadeOutEffect struct {
	duration time.Duration
	startAt  time.Duration
}

// FadeOut creates a video fade-out effect. The start time is calculated
// automatically from the clip's duration during compilation.
func FadeOut(d time.Duration) *FadeOutEffect {
	return &FadeOutEffect{duration: d}
}

// FadeOutAt creates a video fade-out effect starting at a specific time.
func FadeOutAt(start, duration time.Duration) *FadeOutEffect {
	return &FadeOutEffect{duration: duration, startAt: start}
}

func (e *FadeOutEffect) Name() string       { return "fade_out" }
func (e *FadeOutEffect) Target() Target     { return TargetVideo }
func (e *FadeOutEffect) FilterName() string { return "fade" }
func (e *FadeOutEffect) FilterParams() map[string]string {
	return map[string]string{
		"t":  "out",
		"st": formatSeconds(e.startAt),
		"d":  formatSeconds(e.duration),
	}
}

// Dur returns the fade duration.
func (e *FadeOutEffect) Dur() time.Duration { return e.duration }

// StartAt returns when the fade begins (0 means "calculate from clip end").
func (e *FadeOutEffect) StartAt() time.Duration { return e.startAt }

// SpeedEffect changes the playback speed of video.
type SpeedEffect struct {
	factor float64
}

// SpeedUp creates a speed-up effect. Factor > 1 speeds up, < 1 slows down.
//
// Example:
//
//	effects.SpeedUp(2.0)  // 2x speed (timelapse-style)
//	effects.SpeedUp(0.5)  // half speed (slow motion)
func SpeedUp(factor float64) *SpeedEffect {
	return &SpeedEffect{factor: factor}
}

// SlowDown creates a slow-motion effect. Factor is the slowdown amount.
//
// Example:
//
//	effects.SlowDown(2.0)  // 2x slower (same as SpeedUp(0.5))
func SlowDown(factor float64) *SpeedEffect {
	if factor != 0 {
		return &SpeedEffect{factor: 1.0 / factor}
	}
	return &SpeedEffect{factor: 1.0}
}

func (e *SpeedEffect) Name() string       { return "speed" }
func (e *SpeedEffect) Target() Target     { return TargetVideo }
func (e *SpeedEffect) FilterName() string { return "setpts" }
func (e *SpeedEffect) FilterParams() map[string]string {
	// PTS/factor: factor > 1 speeds up (fewer PTS per frame), < 1 slows down.
	return map[string]string{
		"expr": formatFloat(1.0/e.factor) + "*PTS",
	}
}

// Factor returns the speed multiplier.
func (e *SpeedEffect) Factor() float64 { return e.factor }

// --- Audio Effects ---

// AudioFadeInEffect fades audio in from silence.
type AudioFadeInEffect struct {
	duration time.Duration
}

// AudioFadeIn creates an audio fade-in effect.
func AudioFadeIn(d time.Duration) *AudioFadeInEffect {
	return &AudioFadeInEffect{duration: d}
}

func (e *AudioFadeInEffect) Name() string       { return "audio_fade_in" }
func (e *AudioFadeInEffect) Target() Target     { return TargetAudio }
func (e *AudioFadeInEffect) FilterName() string { return "afade" }
func (e *AudioFadeInEffect) FilterParams() map[string]string {
	return map[string]string{
		"t":  "in",
		"st": "0",
		"d":  formatSeconds(e.duration),
	}
}

// Dur returns the fade duration.
func (e *AudioFadeInEffect) Dur() time.Duration { return e.duration }

// AudioFadeOutEffect fades audio to silence.
type AudioFadeOutEffect struct {
	duration time.Duration
	startAt  time.Duration
}

// AudioFadeOut creates an audio fade-out effect.
func AudioFadeOut(d time.Duration) *AudioFadeOutEffect {
	return &AudioFadeOutEffect{duration: d}
}

// AudioFadeOutAt creates an audio fade-out starting at a specific time.
func AudioFadeOutAt(start, duration time.Duration) *AudioFadeOutEffect {
	return &AudioFadeOutEffect{duration: duration, startAt: start}
}

func (e *AudioFadeOutEffect) Name() string       { return "audio_fade_out" }
func (e *AudioFadeOutEffect) Target() Target     { return TargetAudio }
func (e *AudioFadeOutEffect) FilterName() string { return "afade" }
func (e *AudioFadeOutEffect) FilterParams() map[string]string {
	return map[string]string{
		"t":  "out",
		"st": formatSeconds(e.startAt),
		"d":  formatSeconds(e.duration),
	}
}

// Dur returns the fade duration.
func (e *AudioFadeOutEffect) Dur() time.Duration { return e.duration }

// VolumeEffect adjusts the audio volume level.
type VolumeEffect struct {
	level float64
}

// Volume creates a volume adjustment effect.
// 1.0 is original, 0.5 is half, 2.0 is double, 0.0 is mute.
func Volume(level float64) *VolumeEffect {
	return &VolumeEffect{level: level}
}

func (e *VolumeEffect) Name() string       { return "volume" }
func (e *VolumeEffect) Target() Target     { return TargetAudio }
func (e *VolumeEffect) FilterName() string { return "volume" }
func (e *VolumeEffect) FilterParams() map[string]string {
	return map[string]string{
		"volume": formatFloat(e.level),
	}
}

// Level returns the volume multiplier.
func (e *VolumeEffect) Level() float64 { return e.level }

// NormalizeEffect normalizes audio levels using FFmpeg's loudnorm filter.
type NormalizeEffect struct {
	targetLUFS float64
}

// Normalize creates an audio normalization effect targeting -16 LUFS (broadcast standard).
func Normalize() *NormalizeEffect {
	return &NormalizeEffect{targetLUFS: -16.0}
}

// NormalizeTo creates an audio normalization effect targeting a specific LUFS level.
func NormalizeTo(lufs float64) *NormalizeEffect {
	return &NormalizeEffect{targetLUFS: lufs}
}

func (e *NormalizeEffect) Name() string       { return "normalize" }
func (e *NormalizeEffect) Target() Target     { return TargetAudio }
func (e *NormalizeEffect) FilterName() string { return "loudnorm" }
func (e *NormalizeEffect) FilterParams() map[string]string {
	return map[string]string{
		"I": formatFloat(e.targetLUFS),
	}
}

// TargetLUFS returns the target loudness level.
func (e *NormalizeEffect) TargetLUFS() float64 { return e.targetLUFS }

// --- Audio Speed ---

// AudioSpeedEffect changes the audio playback speed (and pitch) using atempo.
type AudioSpeedEffect struct {
	factor float64
}

// AudioSpeed creates an audio speed change effect.
// Factor > 1.0 speeds up, < 1.0 slows down. Range: 0.5 to 100.0.
func AudioSpeed(factor float64) *AudioSpeedEffect {
	return &AudioSpeedEffect{factor: factor}
}

func (e *AudioSpeedEffect) Name() string       { return "audio_speed" }
func (e *AudioSpeedEffect) Target() Target     { return TargetAudio }
func (e *AudioSpeedEffect) FilterName() string { return "atempo" }
func (e *AudioSpeedEffect) FilterParams() map[string]string {
	return map[string]string{
		"tempo": formatFloat(e.factor),
	}
}

// Factor returns the speed multiplier.
func (e *AudioSpeedEffect) Factor() float64 { return e.factor }

// --- Helpers ---

func formatSeconds(d time.Duration) string {
	return formatFloat(d.Seconds())
}

func formatFloat(f float64) string {
	// Use a format that avoids trailing zeros but keeps precision.
	s := fmt.Sprintf("%.6f", f)
	// Trim trailing zeros after decimal point.
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}
