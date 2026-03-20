package cuts

import (
	"time"

	"github.com/ahmedhodiani/gomontage/timeline"
)

// Ensure all cut types implement timeline.Transition at compile time.
var (
	_ timeline.Transition = (*HardCut)(nil)
	_ timeline.Transition = (*LCutTransition)(nil)
	_ timeline.Transition = (*JCutTransition)(nil)
	_ timeline.Transition = (*DissolveTransition)(nil)
	_ timeline.Transition = (*CrossFadeTransition)(nil)
	_ timeline.Transition = (*JumpCutTransition)(nil)
	_ timeline.Transition = (*DipTransition)(nil)
	_ timeline.Transition = (*WipeTransition)(nil)
)

// HardCut is an instant switch from one clip to the next with no transition
// effect. This is the default cut type used when no transition is specified.
type HardCut struct{}

// Hard creates a hard cut transition (instant switch).
func Hard() *HardCut {
	return &HardCut{}
}

func (c *HardCut) Type() timeline.TransitionType { return timeline.TransitionHardCut }
func (c *HardCut) Duration() time.Duration       { return 0 }

// LCutTransition extends the audio from the outgoing clip so it continues
// playing over the incoming clip's video. Named because the edit point forms
// an "L" shape on the timeline.
//
// Use this when cutting from a speaker to B-roll while keeping the dialogue,
// or to smooth the transition between interview segments.
type LCutTransition struct {
	// overlap is how long the outgoing audio extends into the incoming video.
	overlap time.Duration
}

// LCut creates an L-cut transition with the specified audio overlap duration.
//
// The outgoing clip's audio will continue playing for the overlap duration
// after the video has cut to the next clip.
//
// Example:
//
//	tl.AddTransition(cuts.LCut(2*time.Second), interviewClip, brollClip)
func LCut(overlap time.Duration) *LCutTransition {
	return &LCutTransition{overlap: overlap}
}

func (c *LCutTransition) Type() timeline.TransitionType { return timeline.TransitionLCut }
func (c *LCutTransition) Duration() time.Duration       { return c.overlap }

// Overlap returns the duration the outgoing audio extends into the incoming video.
func (c *LCutTransition) Overlap() time.Duration { return c.overlap }

// JCutTransition starts the incoming clip's audio before its video appears.
// Named because the edit point forms a "J" shape on the timeline.
//
// Use this to build anticipation — the audience hears the next scene before
// seeing it. Common in documentaries and narrative filmmaking.
type JCutTransition struct {
	// overlap is how early the incoming audio starts before the video cut.
	overlap time.Duration
}

// JCut creates a J-cut transition with the specified audio lead-in duration.
//
// The incoming clip's audio will start playing for the overlap duration
// before the video cuts to the incoming clip.
//
// Example:
//
//	tl.AddTransition(cuts.JCut(1*time.Second), scene1, scene2)
func JCut(overlap time.Duration) *JCutTransition {
	return &JCutTransition{overlap: overlap}
}

func (c *JCutTransition) Type() timeline.TransitionType { return timeline.TransitionJCut }
func (c *JCutTransition) Duration() time.Duration       { return c.overlap }

// Overlap returns how early the incoming audio starts before the video cut.
func (c *JCutTransition) Overlap() time.Duration { return c.overlap }

// DissolveTransition crossfades the video between two clips. The outgoing
// clip fades out while the incoming clip fades in simultaneously.
type DissolveTransition struct {
	duration time.Duration
}

// Dissolve creates a dissolve (video crossfade) transition.
//
// Example:
//
//	tl.AddTransition(cuts.Dissolve(1*time.Second), clip1, clip2)
func Dissolve(d time.Duration) *DissolveTransition {
	return &DissolveTransition{duration: d}
}

func (c *DissolveTransition) Type() timeline.TransitionType { return timeline.TransitionDissolve }
func (c *DissolveTransition) Duration() time.Duration       { return c.duration }

// CrossFadeTransition crossfades both video AND audio between two clips.
// This is a dissolve for video combined with an audio crossfade.
type CrossFadeTransition struct {
	duration time.Duration
}

// CrossFade creates a crossfade transition for both video and audio.
//
// Example:
//
//	tl.AddTransition(cuts.CrossFade(1*time.Second), clip1, clip2)
func CrossFade(d time.Duration) *CrossFadeTransition {
	return &CrossFadeTransition{duration: d}
}

func (c *CrossFadeTransition) Type() timeline.TransitionType { return timeline.TransitionCrossFade }
func (c *CrossFadeTransition) Duration() time.Duration       { return c.duration }

// JumpCutTransition removes a section of footage and snaps the remaining
// clips together with no blending. Creates an intentional jarring effect
// often used in vlogs and YouTube content.
type JumpCutTransition struct{}

// JumpCut creates a jump cut transition (sharp, jarring cut).
//
// Example:
//
//	tl.AddTransition(cuts.JumpCut(), beforePause, afterPause)
func JumpCut() *JumpCutTransition {
	return &JumpCutTransition{}
}

func (c *JumpCutTransition) Type() timeline.TransitionType { return timeline.TransitionJumpCut }
func (c *JumpCutTransition) Duration() time.Duration       { return 0 }

// DipTransition fades the outgoing clip to a solid color, then fades in
// the incoming clip from that color. The total transition takes twice the
// specified duration (fade out + fade in).
type DipTransition struct {
	duration  time.Duration
	color     string // "black" or "white"
	transType timeline.TransitionType
}

// DipToBlack creates a dip-to-black transition.
// The outgoing clip fades to black, then the incoming clip fades in from black.
//
// Example:
//
//	tl.AddTransition(cuts.DipToBlack(1*time.Second), scene1, scene2)
func DipToBlack(d time.Duration) *DipTransition {
	return &DipTransition{
		duration:  d,
		color:     "black",
		transType: timeline.TransitionDipToBlack,
	}
}

// DipToWhite creates a dip-to-white transition.
// The outgoing clip fades to white, then the incoming clip fades in from white.
//
// Example:
//
//	tl.AddTransition(cuts.DipToWhite(500*time.Millisecond), clip1, clip2)
func DipToWhite(d time.Duration) *DipTransition {
	return &DipTransition{
		duration:  d,
		color:     "white",
		transType: timeline.TransitionDipToWhite,
	}
}

func (c *DipTransition) Type() timeline.TransitionType { return c.transType }
func (c *DipTransition) Duration() time.Duration       { return c.duration * 2 } // Fade out + fade in.

// Color returns the dip color ("black" or "white").
func (c *DipTransition) Color() string { return c.color }

// HalfDuration returns the duration of each half (fade out or fade in).
func (c *DipTransition) HalfDuration() time.Duration { return c.duration }

// WipeDirection specifies the direction of a wipe transition.
type WipeDirection int

const (
	WipeLeft WipeDirection = iota
	WipeRight
	WipeUp
	WipeDown
)

// String returns the direction as a human-readable string.
func (d WipeDirection) String() string {
	switch d {
	case WipeLeft:
		return "left"
	case WipeRight:
		return "right"
	case WipeUp:
		return "up"
	case WipeDown:
		return "down"
	default:
		return "unknown"
	}
}

// WipeTransition reveals the incoming clip by moving a boundary across the frame.
type WipeTransition struct {
	duration  time.Duration
	direction WipeDirection
}

// Wipe creates a wipe transition in the specified direction.
//
// Example:
//
//	tl.AddTransition(cuts.Wipe(cuts.WipeLeft, 1*time.Second), clip1, clip2)
func Wipe(direction WipeDirection, d time.Duration) *WipeTransition {
	return &WipeTransition{
		duration:  d,
		direction: direction,
	}
}

func (c *WipeTransition) Type() timeline.TransitionType { return timeline.TransitionWipe }
func (c *WipeTransition) Duration() time.Duration       { return c.duration }

// Direction returns the wipe direction.
func (c *WipeTransition) Direction() WipeDirection { return c.direction }
