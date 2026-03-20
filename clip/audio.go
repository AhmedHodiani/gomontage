package clip

import (
	"time"

	"github.com/ahmedhodiani/gomontage/engine"
)

// AudioClip represents a clip loaded from an audio file (wav, mp3, flac, etc.).
type AudioClip struct {
	Base
}

// NewAudio creates a new AudioClip from an audio file path.
//
// The file is probed to extract metadata (duration, sample rate, channels, etc.).
//
// Example:
//
//	narration := clip.NewAudio("resources/audio/narration.wav")
//	music := clip.NewAudio("resources/audio/bgm.mp3").WithVolume(0.3)
func NewAudio(path string) *AudioClip {
	c := &AudioClip{
		Base: Base{
			clipType:   TypeAudio,
			sourcePath: path,
			volume:     1.0,
			hasAudio:   true,
		},
	}

	// Probe the file for metadata.
	info, err := engine.Probe(path)
	if err == nil {
		c.duration = info.Duration
		c.trimEnd = info.Duration
	}

	return c
}

// NewAudioWithDuration creates an AudioClip with a manually specified duration,
// skipping the file probe.
func NewAudioWithDuration(path string, duration time.Duration) *AudioClip {
	return &AudioClip{
		Base: Base{
			clipType:   TypeAudio,
			sourcePath: path,
			duration:   duration,
			trimEnd:    duration,
			volume:     1.0,
			hasAudio:   true,
		},
	}
}

// Trim returns a new AudioClip that only includes the segment from start to end.
//
// Example:
//
//	full := clip.NewAudio("narration.wav")             // 5 minute narration
//	intro := full.Trim(0, 30*time.Second)              // first 30s
//	chapter := full.Trim(time.Minute, 2*time.Minute)   // 1 minute segment
func (c *AudioClip) Trim(start, end time.Duration) *AudioClip {
	n := &AudioClip{Base: *c.base()}
	n.trimStart = start
	n.trimEnd = end
	n.duration = end - start
	n.trimmed = true
	return n
}

// WithVolume returns a new AudioClip with the volume adjusted.
// 1.0 is original volume, 0.5 is half, 2.0 is double, 0.0 is mute.
//
// Example:
//
//	bgm := clip.NewAudio("music.mp3").WithVolume(0.2) // Quiet background music
func (c *AudioClip) WithVolume(vol float64) *AudioClip {
	n := &AudioClip{Base: *c.base()}
	n.volume = vol
	return n
}

// WithFadeIn returns a new AudioClip with a fade-in at the start.
//
// Example:
//
//	music := clip.NewAudio("music.mp3").WithFadeIn(3 * time.Second)
func (c *AudioClip) WithFadeIn(d time.Duration) *AudioClip {
	n := &AudioClip{Base: *c.base()}
	n.fadeIn = d
	return n
}

// WithFadeOut returns a new AudioClip with a fade-out at the end.
//
// Example:
//
//	music := clip.NewAudio("music.mp3").WithFadeOut(5 * time.Second)
func (c *AudioClip) WithFadeOut(d time.Duration) *AudioClip {
	n := &AudioClip{Base: *c.base()}
	n.fadeOut = d
	return n
}

// WithDuration returns a new AudioClip with a specific duration.
// If shorter than the original, it trims from the start. If longer, behavior
// depends on the audio (typically silence after the end).
func (c *AudioClip) WithDuration(d time.Duration) *AudioClip {
	n := &AudioClip{Base: *c.base()}
	n.duration = d
	n.trimEnd = n.trimStart + d
	return n
}
