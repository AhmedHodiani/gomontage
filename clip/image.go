package clip

import (
	"time"

	"github.com/ahmedhodiani/gomontage/effects"
)

// ImageClip represents a static image used as a video clip.
// Images have infinite natural duration, so you must set a duration
// either directly or through placement on a timeline.
type ImageClip struct {
	Base
}

// NewImage creates a new ImageClip from an image file path.
// The default duration is 5 seconds — use WithDuration to change it.
//
// Example:
//
//	logo := clip.NewImage("resources/images/logo.png").WithDuration(3 * time.Second)
//	bg := clip.NewImage("resources/images/background.jpg")
func NewImage(path string) *ImageClip {
	return &ImageClip{
		Base: Base{
			clipType:   TypeImage,
			sourcePath: path,
			duration:   5 * time.Second, // Default duration for images.
			trimEnd:    5 * time.Second,
			hasVideo:   true,
			volume:     1.0,
		},
	}
}

// WithDuration returns a new ImageClip with the specified duration.
func (c *ImageClip) WithDuration(d time.Duration) *ImageClip {
	n := &ImageClip{Base: *c.base()}
	n.duration = d
	n.trimEnd = d
	return n
}

// WithPosition returns a new ImageClip placed at the given position.
func (c *ImageClip) WithPosition(pos Position) *ImageClip {
	n := &ImageClip{Base: *c.base()}
	n.position = pos
	return n
}

// WithSize returns a new ImageClip scaled to the given dimensions.
func (c *ImageClip) WithSize(width, height int) *ImageClip {
	n := &ImageClip{Base: *c.base()}
	n.width = width
	n.height = height
	return n
}

// WithFadeIn returns a new ImageClip with a fade-in at the start.
func (c *ImageClip) WithFadeIn(d time.Duration) *ImageClip {
	n := &ImageClip{Base: *c.base()}
	n.fadeIn = d
	return n
}

// WithFadeOut returns a new ImageClip with a fade-out at the end.
func (c *ImageClip) WithFadeOut(d time.Duration) *ImageClip {
	n := &ImageClip{Base: *c.base()}
	n.fadeOut = d
	return n
}

// WithEffect returns a new ImageClip with the given effect appended.
// Effects are composable — call WithEffect multiple times to stack effects.
//
// Example:
//
//	clip.NewImage("logo.png").
//	    WithDuration(5 * time.Second).
//	    WithEffect(effects.FadeIn(1 * time.Second))
func (c *ImageClip) WithEffect(e effects.Effect) *ImageClip {
	n := &ImageClip{Base: *c.base()}
	n.effects = append(n.effects, e)
	return n
}
