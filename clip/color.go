package clip

import "time"

// ColorClip represents a solid color rectangle, useful for backgrounds,
// letterboxing, or color fills between clips.
type ColorClip struct {
	Base

	// Color is the fill color as a hex string (e.g. "#000000" for black).
	Color string
}

// NewColor creates a new ColorClip with the specified color and dimensions.
// The default duration is 5 seconds — use WithDuration to change it.
//
// Example:
//
//	black := clip.NewColor("#000000", 1920, 1080).WithDuration(2 * time.Second)
//	red := clip.NewColor("#FF0000", 1920, 1080)
func NewColor(color string, width, height int) *ColorClip {
	return &ColorClip{
		Base: Base{
			clipType: TypeColor,
			duration: 5 * time.Second,
			trimEnd:  5 * time.Second,
			hasVideo: true,
			volume:   1.0,
			width:    width,
			height:   height,
		},
		Color: color,
	}
}

// WithDuration returns a new ColorClip with the specified duration.
func (c *ColorClip) WithDuration(d time.Duration) *ColorClip {
	n := &ColorClip{Base: *c.base(), Color: c.Color}
	n.duration = d
	n.trimEnd = d
	return n
}

// WithFadeIn returns a new ColorClip with a fade-in effect.
func (c *ColorClip) WithFadeIn(d time.Duration) *ColorClip {
	n := &ColorClip{Base: *c.base(), Color: c.Color}
	n.fadeIn = d
	return n
}

// WithFadeOut returns a new ColorClip with a fade-out effect.
func (c *ColorClip) WithFadeOut(d time.Duration) *ColorClip {
	n := &ColorClip{Base: *c.base(), Color: c.Color}
	n.fadeOut = d
	return n
}
