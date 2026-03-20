// Package gomontage is a programmatic video editing framework for Go.
//
// Gomontage lets you describe video edits with high-level Go code using a
// track-based timeline model. Load clips, arrange them on tracks, add
// effects, layer audio, and export — all in clean,
// readable Go.
//
// # Quick Start
//
//	tl := gomontage.NewTimeline(gomontage.TimelineConfig{
//	    Width:  1920,
//	    Height: 1080,
//	    FPS:    30,
//	})
//
//	video := tl.AddVideoTrack("main")
//	video.Add(clip.NewVideo("intro.mp4"), gomontage.At(0))
//
//	tl.Export(export.YouTube1080p(), "output/final.mp4")
//
// See the subpackages for detailed documentation:
//   - clip: Media clip types (video, audio, image, text, color)
//   - timeline: Track-based timeline and composition
//   - effects: Audio and video effects (fade, volume, speed)
//   - export: Output profiles and presets
//   - engine: Low-level FFmpeg interface (internal)
package gomontage

import (
	"time"

	"github.com/ahmedhodiani/gomontage/timeline"
)

// TimelineConfig is an alias for timeline.Config for convenience at the
// top-level package.
type TimelineConfig = timeline.Config

// NewTimeline creates a new Timeline with the given configuration.
// This is a convenience wrapper around timeline.New.
//
// Example:
//
//	tl := gomontage.NewTimeline(gomontage.TimelineConfig{
//	    Width:  1920,
//	    Height: 1080,
//	    FPS:    30,
//	})
func NewTimeline(cfg TimelineConfig) *timeline.Timeline {
	return timeline.New(cfg)
}

// At returns a time.Duration for placing a clip at a specific position
// on a track. This is a convenience wrapper around timeline.At.
//
// Example:
//
//	track.Add(myClip, gomontage.At(5*time.Second))
func At(t time.Duration) time.Duration {
	return timeline.At(t)
}

// HD returns a TimelineConfig for 1920x1080 at 30fps.
func HD() TimelineConfig {
	return TimelineConfig{
		Width:  1920,
		Height: 1080,
		FPS:    30,
	}
}

// HD60 returns a TimelineConfig for 1920x1080 at 60fps.
func HD60() TimelineConfig {
	return TimelineConfig{
		Width:  1920,
		Height: 1080,
		FPS:    60,
	}
}

// UHD returns a TimelineConfig for 3840x2160 (4K) at 30fps.
func UHD() TimelineConfig {
	return TimelineConfig{
		Width:  3840,
		Height: 2160,
		FPS:    30,
	}
}

// UHD60 returns a TimelineConfig for 3840x2160 (4K) at 60fps.
func UHD60() TimelineConfig {
	return TimelineConfig{
		Width:  3840,
		Height: 2160,
		FPS:    60,
	}
}

// Vertical returns a TimelineConfig for 1080x1920 (9:16 vertical) at 30fps.
// Suitable for Instagram Reels, TikTok, YouTube Shorts.
func Vertical() TimelineConfig {
	return TimelineConfig{
		Width:  1080,
		Height: 1920,
		FPS:    30,
	}
}

// Square returns a TimelineConfig for 1080x1080 (1:1 square) at 30fps.
func Square() TimelineConfig {
	return TimelineConfig{
		Width:  1080,
		Height: 1080,
		FPS:    30,
	}
}

// Seconds is a convenience function that converts a float64 to time.Duration.
// This makes clip timing expressions more readable.
//
// Example:
//
//	clip.Trim(gomontage.Seconds(5.5), gomontage.Seconds(30.0))
func Seconds(s float64) time.Duration {
	return time.Duration(s * float64(time.Second))
}

// Minutes is a convenience function that converts a float64 to time.Duration.
//
// Example:
//
//	clip.Trim(gomontage.Minutes(1), gomontage.Minutes(2.5))
func Minutes(m float64) time.Duration {
	return time.Duration(m * float64(time.Minute))
}
