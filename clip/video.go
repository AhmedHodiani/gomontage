package clip

import (
	"time"

	"github.com/ahmedhodiani/gomontage/engine"
)

// VideoClip represents a clip loaded from a video file.
// It may contain both video and audio streams.
type VideoClip struct {
	Base
}

// NewVideo creates a new VideoClip from a video file path.
//
// The file is probed immediately to extract metadata (duration, resolution, etc.).
// Returns a VideoClip with the full file duration and original properties.
//
// Example:
//
//	interview := clip.NewVideo("resources/video/interview.mp4")
//	trimmed := interview.Trim(10*time.Second, 30*time.Second)
func NewVideo(path string) *VideoClip {
	c := &VideoClip{
		Base: Base{
			clipType:   TypeVideo,
			sourcePath: path,
			volume:     1.0,
			hasVideo:   true,
		},
	}

	// Probe the file for metadata.
	info, err := engine.Probe(path)
	if err == nil {
		c.duration = info.Duration
		c.trimEnd = info.Duration
		c.hasAudio = info.HasAudio()
		if len(info.VideoStreams) > 0 {
			c.width = info.VideoStreams[0].Width
			c.height = info.VideoStreams[0].Height
		}
	}

	return c
}

// NewVideoWithDuration creates a VideoClip with a manually specified duration,
// skipping the file probe. Useful for testing or when you know the duration.
func NewVideoWithDuration(path string, duration time.Duration) *VideoClip {
	return &VideoClip{
		Base: Base{
			clipType:   TypeVideo,
			sourcePath: path,
			duration:   duration,
			trimEnd:    duration,
			volume:     1.0,
			hasVideo:   true,
			hasAudio:   true,
		},
	}
}

// Trim returns a new VideoClip that only includes the segment from start to end.
// Times are relative to the original source file.
//
// Example:
//
//	full := clip.NewVideo("interview.mp4")       // 60s video
//	intro := full.Trim(0, 10*time.Second)        // first 10s
//	middle := full.Trim(20*time.Second, 40*time.Second) // 20s segment
func (c *VideoClip) Trim(start, end time.Duration) *VideoClip {
	n := &VideoClip{Base: *c.base()}
	n.trimStart = start
	n.trimEnd = end
	n.duration = end - start
	n.trimmed = true
	return n
}

// WithVolume returns a new VideoClip with the audio volume adjusted.
// 1.0 is original volume, 0.5 is half, 2.0 is double.
// Use 0.0 to mute the audio.
func (c *VideoClip) WithVolume(vol float64) *VideoClip {
	n := &VideoClip{Base: *c.base()}
	n.volume = vol
	return n
}

// WithFadeIn returns a new VideoClip with a fade-in effect applied to both
// video (opacity) and audio (volume) at the start.
func (c *VideoClip) WithFadeIn(d time.Duration) *VideoClip {
	n := &VideoClip{Base: *c.base()}
	n.fadeIn = d
	return n
}

// WithFadeOut returns a new VideoClip with a fade-out effect applied to both
// video and audio at the end.
func (c *VideoClip) WithFadeOut(d time.Duration) *VideoClip {
	n := &VideoClip{Base: *c.base()}
	n.fadeOut = d
	return n
}

// WithPosition returns a new VideoClip placed at the given position.
// Used when this clip is on an overlay track.
func (c *VideoClip) WithPosition(pos Position) *VideoClip {
	n := &VideoClip{Base: *c.base()}
	n.position = pos
	return n
}

// WithSize returns a new VideoClip scaled to the given dimensions.
func (c *VideoClip) WithSize(width, height int) *VideoClip {
	n := &VideoClip{Base: *c.base()}
	n.width = width
	n.height = height
	return n
}

// AudioOnly returns a new VideoClip with video disabled, keeping only audio.
// Useful when you want to use a video file's audio track separately.
func (c *VideoClip) AudioOnly() *VideoClip {
	n := &VideoClip{Base: *c.base()}
	n.hasVideo = false
	return n
}

// VideoOnly returns a new VideoClip with audio disabled, keeping only video.
// Useful when you want to replace the audio with narration or music.
func (c *VideoClip) VideoOnly() *VideoClip {
	n := &VideoClip{Base: *c.base()}
	n.hasAudio = false
	return n
}
