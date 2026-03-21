package clip

import (
	"time"

	"github.com/ahmedhodiani/gomontage/effects"
)

// TextStyle configures the visual appearance of a TextClip.
type TextStyle struct {
	// Font is the path to a .ttf or .otf font file.
	// If empty, FFmpeg's default font is used.
	Font string

	// Size is the font size in points. Default is 48.
	Size int

	// Color is the text color as a hex string (e.g. "#FFFFFF") or named color.
	// Default is white.
	Color string

	// Background is the background color behind the text. Empty means transparent.
	Background string

	// BorderWidth is the width of the text border/outline in pixels. 0 = no border.
	BorderWidth int

	// BorderColor is the border color. Default is black.
	BorderColor string

	// Position is where the text is placed on the canvas.
	Position Position

	// BoxPadding is the padding around the text when Background is set.
	BoxPadding int
}

// DefaultTextStyle returns a TextStyle with sensible defaults.
func DefaultTextStyle() TextStyle {
	return TextStyle{
		Size:        48,
		Color:       "#FFFFFF",
		BorderColor: "#000000",
		BorderWidth: 2,
	}
}

// TextClip represents a dynamically generated text overlay.
type TextClip struct {
	Base

	// Text is the string content to render.
	Text string

	// Style controls the visual appearance.
	Style TextStyle
}

// NewText creates a new TextClip with the given text and style.
// The default duration is 5 seconds — use WithDuration to change it.
//
// Example:
//
//	title := clip.NewText("Chapter 1: The Beginning", clip.TextStyle{
//	    Font:  "resources/fonts/bold.ttf",
//	    Size:  72,
//	    Color: "#FFFFFF",
//	}).WithDuration(4 * time.Second)
func NewText(text string, style TextStyle) *TextClip {
	if style.Size == 0 {
		style.Size = DefaultTextStyle().Size
	}
	if style.Color == "" {
		style.Color = DefaultTextStyle().Color
	}
	return &TextClip{
		Base: Base{
			clipType: TypeText,
			duration: 5 * time.Second,
			trimEnd:  5 * time.Second,
			hasVideo: true,
			volume:   1.0,
			position: style.Position,
		},
		Text:  text,
		Style: style,
	}
}

// WithDuration returns a new TextClip with the specified duration.
func (c *TextClip) WithDuration(d time.Duration) *TextClip {
	n := &TextClip{Base: *c.base(), Text: c.Text, Style: c.Style}
	n.duration = d
	n.trimEnd = d
	return n
}

// WithPosition returns a new TextClip placed at the given position.
func (c *TextClip) WithPosition(pos Position) *TextClip {
	n := &TextClip{Base: *c.base(), Text: c.Text, Style: c.Style}
	n.position = pos
	return n
}

// WithFadeIn returns a new TextClip with a fade-in effect.
func (c *TextClip) WithFadeIn(d time.Duration) *TextClip {
	n := &TextClip{Base: *c.base(), Text: c.Text, Style: c.Style}
	n.fadeIn = d
	return n
}

// WithFadeOut returns a new TextClip with a fade-out effect.
func (c *TextClip) WithFadeOut(d time.Duration) *TextClip {
	n := &TextClip{Base: *c.base(), Text: c.Text, Style: c.Style}
	n.fadeOut = d
	return n
}

// WithEffect returns a new TextClip with the given effect appended.
// Effects are composable — call WithEffect multiple times to stack effects.
//
// Example:
//
//	clip.NewText("Chapter 1", clip.DefaultTextStyle()).
//	    WithDuration(4 * time.Second).
//	    WithEffect(effects.FadeIn(1 * time.Second))
func (c *TextClip) WithEffect(e effects.Effect) *TextClip {
	n := &TextClip{Base: *c.base(), Text: c.Text, Style: c.Style}
	n.applyEffect(e)
	return n
}
