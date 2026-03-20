package clip

import "time"

// Type identifies the kind of media a clip represents.
type Type int

const (
	// TypeVideo is a clip loaded from a video file (may contain audio).
	TypeVideo Type = iota
	// TypeAudio is a clip loaded from an audio-only file.
	TypeAudio
	// TypeImage is a clip loaded from a static image file.
	TypeImage
	// TypeText is a dynamically generated text overlay.
	TypeText
	// TypeColor is a solid color fill.
	TypeColor
)

// String returns a human-readable name for the clip type.
func (t Type) String() string {
	switch t {
	case TypeVideo:
		return "video"
	case TypeAudio:
		return "audio"
	case TypeImage:
		return "image"
	case TypeText:
		return "text"
	case TypeColor:
		return "color"
	default:
		return "unknown"
	}
}

// Position specifies where a clip is placed spatially (for overlays).
type Position struct {
	// X is the horizontal position in pixels from the left.
	X int
	// Y is the vertical position in pixels from the top.
	Y int
	// Relative indicates if X and Y are relative to the canvas (0.0-1.0).
	Relative bool
}

// Predefined positions for common placements.
var (
	// Center places the clip at the center of the canvas.
	Center = Position{X: 0, Y: 0, Relative: false} // Resolved by the compiler.
	// TopLeft places the clip at the top-left corner.
	TopLeft = Position{X: 0, Y: 0}
	// TopRight places the clip at the top-right corner (resolved by compiler).
	TopRight = Position{X: -1, Y: 0}
	// BottomLeft places the clip at the bottom-left corner (resolved by compiler).
	BottomLeft = Position{X: 0, Y: -1}
	// BottomRight places the clip at the bottom-right corner (resolved by compiler).
	BottomRight = Position{X: -1, Y: -1}
)

// Clip is the core interface that all clip types implement.
// Clips are immutable — every modification method returns a new Clip.
type Clip interface {
	// ClipType returns what kind of clip this is (video, audio, image, text, color).
	ClipType() Type

	// Duration returns the clip's duration. For trimmed clips, this is the
	// trimmed duration, not the original file duration.
	Duration() time.Duration

	// TrimStart returns the start time of the trim window within the source.
	TrimStart() time.Duration

	// TrimEnd returns the end time of the trim window within the source.
	TrimEnd() time.Duration

	// SourcePath returns the path to the source file, or empty for generated clips.
	SourcePath() string

	// HasVideo returns true if this clip contains a video stream.
	HasVideo() bool

	// HasAudio returns true if this clip contains an audio stream.
	HasAudio() bool

	// Volume returns the audio volume multiplier (1.0 = original).
	Volume() float64

	// FadeInDuration returns the fade-in duration (0 = no fade).
	FadeInDuration() time.Duration

	// FadeOutDuration returns the fade-out duration (0 = no fade).
	FadeOutDuration() time.Duration

	// Pos returns the spatial position for overlay clips.
	Pos() Position

	// Width returns the clip width in pixels (0 if unknown or audio-only).
	Width() int

	// Height returns the clip height in pixels (0 if unknown or audio-only).
	Height() int

	// clone returns a deep copy of the clip's base properties.
	// This is internal — used by transform methods to create modified copies.
	base() *Base
}

// Base holds the common properties shared by all clip types.
// It is embedded in every concrete clip type.
type Base struct {
	clipType   Type
	sourcePath string
	duration   time.Duration
	trimStart  time.Duration
	trimEnd    time.Duration
	hasVideo   bool
	hasAudio   bool
	volume     float64
	fadeIn     time.Duration
	fadeOut    time.Duration
	position   Position
	width      int
	height     int
}

func (b *Base) ClipType() Type                 { return b.clipType }
func (b *Base) Duration() time.Duration        { return b.duration }
func (b *Base) TrimStart() time.Duration       { return b.trimStart }
func (b *Base) TrimEnd() time.Duration         { return b.trimEnd }
func (b *Base) SourcePath() string             { return b.sourcePath }
func (b *Base) HasVideo() bool                 { return b.hasVideo }
func (b *Base) HasAudio() bool                 { return b.hasAudio }
func (b *Base) Volume() float64                { return b.volume }
func (b *Base) FadeInDuration() time.Duration  { return b.fadeIn }
func (b *Base) FadeOutDuration() time.Duration { return b.fadeOut }
func (b *Base) Pos() Position                  { return b.position }
func (b *Base) Width() int                     { return b.width }
func (b *Base) Height() int                    { return b.height }

func (b *Base) base() *Base {
	// Return a shallow copy.
	copy := *b
	return &copy
}
