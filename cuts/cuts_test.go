package cuts

import (
	"testing"
	"time"

	"github.com/ahmedhodiani/gomontage/timeline"
)

func TestHardCut(t *testing.T) {
	c := Hard()
	if c.Type() != timeline.TransitionHardCut {
		t.Errorf("expected HardCut type, got %v", c.Type())
	}
	if c.Duration() != 0 {
		t.Errorf("hard cut should have 0 duration, got %v", c.Duration())
	}
}

func TestLCut(t *testing.T) {
	c := LCut(2 * time.Second)
	if c.Type() != timeline.TransitionLCut {
		t.Errorf("expected LCut type, got %v", c.Type())
	}
	if c.Duration() != 2*time.Second {
		t.Errorf("expected 2s, got %v", c.Duration())
	}
	if c.Overlap() != 2*time.Second {
		t.Errorf("expected overlap 2s, got %v", c.Overlap())
	}
}

func TestJCut(t *testing.T) {
	c := JCut(1 * time.Second)
	if c.Type() != timeline.TransitionJCut {
		t.Errorf("expected JCut type, got %v", c.Type())
	}
	if c.Duration() != 1*time.Second {
		t.Errorf("expected 1s, got %v", c.Duration())
	}
	if c.Overlap() != 1*time.Second {
		t.Errorf("expected overlap 1s, got %v", c.Overlap())
	}
}

func TestDissolve(t *testing.T) {
	c := Dissolve(1 * time.Second)
	if c.Type() != timeline.TransitionDissolve {
		t.Errorf("expected Dissolve type, got %v", c.Type())
	}
	if c.Duration() != 1*time.Second {
		t.Errorf("expected 1s, got %v", c.Duration())
	}
}

func TestCrossFade(t *testing.T) {
	c := CrossFade(500 * time.Millisecond)
	if c.Type() != timeline.TransitionCrossFade {
		t.Errorf("expected CrossFade type, got %v", c.Type())
	}
	if c.Duration() != 500*time.Millisecond {
		t.Errorf("expected 500ms, got %v", c.Duration())
	}
}

func TestJumpCut(t *testing.T) {
	c := JumpCut()
	if c.Type() != timeline.TransitionJumpCut {
		t.Errorf("expected JumpCut type, got %v", c.Type())
	}
	if c.Duration() != 0 {
		t.Errorf("jump cut should have 0 duration, got %v", c.Duration())
	}
}

func TestDipToBlack(t *testing.T) {
	c := DipToBlack(1 * time.Second)
	if c.Type() != timeline.TransitionDipToBlack {
		t.Errorf("expected DipToBlack type, got %v", c.Type())
	}
	// Total duration is 2x (fade out + fade in).
	if c.Duration() != 2*time.Second {
		t.Errorf("expected 2s total, got %v", c.Duration())
	}
	if c.Color() != "black" {
		t.Errorf("expected black, got %s", c.Color())
	}
	if c.HalfDuration() != 1*time.Second {
		t.Errorf("expected half 1s, got %v", c.HalfDuration())
	}
}

func TestDipToWhite(t *testing.T) {
	c := DipToWhite(500 * time.Millisecond)
	if c.Type() != timeline.TransitionDipToWhite {
		t.Errorf("expected DipToWhite type, got %v", c.Type())
	}
	if c.Duration() != 1*time.Second {
		t.Errorf("expected 1s total, got %v", c.Duration())
	}
	if c.Color() != "white" {
		t.Errorf("expected white, got %s", c.Color())
	}
}

func TestWipe(t *testing.T) {
	c := Wipe(WipeLeft, 1*time.Second)
	if c.Type() != timeline.TransitionWipe {
		t.Errorf("expected Wipe type, got %v", c.Type())
	}
	if c.Duration() != 1*time.Second {
		t.Errorf("expected 1s, got %v", c.Duration())
	}
	if c.Direction() != WipeLeft {
		t.Errorf("expected WipeLeft, got %v", c.Direction())
	}
}

func TestWipeDirection_String(t *testing.T) {
	tests := []struct {
		d    WipeDirection
		want string
	}{
		{WipeLeft, "left"},
		{WipeRight, "right"},
		{WipeUp, "up"},
		{WipeDown, "down"},
		{WipeDirection(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.d.String(); got != tt.want {
			t.Errorf("WipeDirection(%d).String() = %q, want %q", int(tt.d), got, tt.want)
		}
	}
}

func TestInterfaceCompliance(t *testing.T) {
	// This test verifies all types implement the Transition interface.
	// The compile-time checks in cuts.go handle this, but this is explicit.
	transitions := []timeline.Transition{
		Hard(),
		LCut(1 * time.Second),
		JCut(1 * time.Second),
		Dissolve(1 * time.Second),
		CrossFade(1 * time.Second),
		JumpCut(),
		DipToBlack(1 * time.Second),
		DipToWhite(1 * time.Second),
		Wipe(WipeLeft, 1*time.Second),
	}

	for i, tr := range transitions {
		if tr.Type() < 0 {
			t.Errorf("transition %d has invalid type", i)
		}
	}
}
