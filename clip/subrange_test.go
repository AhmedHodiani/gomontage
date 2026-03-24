package clip

import (
	"testing"
	"time"
)

func TestSubRange_VideoClip_FullWindow(t *testing.T) {
	// SubRange covering the full clip should return the same clip.
	v := NewVideoWithDuration("test.mp4", 30*time.Second)
	result := SubRange(v, 0, 30*time.Second)
	if result != v {
		t.Error("SubRange over the full clip should return the original clip")
	}
}

func TestSubRange_VideoClip_FrontTrim(t *testing.T) {
	// Cut the first 10s off a 30s video clip.
	v := NewVideoWithDuration("test.mp4", 30*time.Second)
	result := SubRange(v, 10*time.Second, 30*time.Second)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Duration() != 20*time.Second {
		t.Errorf("expected duration 20s, got %v", result.Duration())
	}
	if result.TrimStart() != 10*time.Second {
		t.Errorf("expected TrimStart 10s, got %v", result.TrimStart())
	}
	if result.TrimEnd() != 30*time.Second {
		t.Errorf("expected TrimEnd 30s, got %v", result.TrimEnd())
	}
	if !result.IsTrimmed() {
		t.Error("expected IsTrimmed=true")
	}
}

func TestSubRange_VideoClip_BackTrim(t *testing.T) {
	// Keep only the first 10s of a 30s video clip.
	v := NewVideoWithDuration("test.mp4", 30*time.Second)
	result := SubRange(v, 0, 10*time.Second)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Duration() != 10*time.Second {
		t.Errorf("expected duration 10s, got %v", result.Duration())
	}
	if result.TrimStart() != 0 {
		t.Errorf("expected TrimStart 0, got %v", result.TrimStart())
	}
	if result.TrimEnd() != 10*time.Second {
		t.Errorf("expected TrimEnd 10s, got %v", result.TrimEnd())
	}
}

func TestSubRange_VideoClip_MiddleTrim(t *testing.T) {
	// Extract seconds 10-20 from a 30s video clip.
	v := NewVideoWithDuration("test.mp4", 30*time.Second)
	result := SubRange(v, 10*time.Second, 20*time.Second)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Duration() != 10*time.Second {
		t.Errorf("expected duration 10s, got %v", result.Duration())
	}
	if result.TrimStart() != 10*time.Second {
		t.Errorf("expected TrimStart 10s, got %v", result.TrimStart())
	}
	if result.TrimEnd() != 20*time.Second {
		t.Errorf("expected TrimEnd 20s, got %v", result.TrimEnd())
	}
}

func TestSubRange_VideoClip_AlreadyTrimmed(t *testing.T) {
	// SubRange on a clip that was already trimmed (source seconds 20-40).
	// The clip-local duration is 20s. SubRange(5s, 15s) should produce
	// source coordinates 25s-35s.
	v := NewVideoWithDuration("test.mp4", 60*time.Second)
	trimmed := v.Trim(20*time.Second, 40*time.Second) // 20s clip, source 20-40

	result := SubRange(trimmed, 5*time.Second, 15*time.Second)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Duration() != 10*time.Second {
		t.Errorf("expected duration 10s, got %v", result.Duration())
	}
	if result.TrimStart() != 25*time.Second {
		t.Errorf("expected TrimStart 25s (source), got %v", result.TrimStart())
	}
	if result.TrimEnd() != 35*time.Second {
		t.Errorf("expected TrimEnd 35s (source), got %v", result.TrimEnd())
	}
}

func TestSubRange_VideoClip_EmptyWindow(t *testing.T) {
	v := NewVideoWithDuration("test.mp4", 30*time.Second)

	// start >= end
	if SubRange(v, 20*time.Second, 10*time.Second) != nil {
		t.Error("expected nil for inverted window")
	}

	// start == end
	if SubRange(v, 10*time.Second, 10*time.Second) != nil {
		t.Error("expected nil for zero-width window")
	}

	// Entirely past the clip
	if SubRange(v, 40*time.Second, 50*time.Second) != nil {
		t.Error("expected nil for window past clip end")
	}
}

func TestSubRange_VideoClip_ClampedWindow(t *testing.T) {
	// Window extends past the clip — should be clamped.
	v := NewVideoWithDuration("test.mp4", 30*time.Second)
	result := SubRange(v, 20*time.Second, 60*time.Second)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Duration() != 10*time.Second {
		t.Errorf("expected duration 10s (clamped), got %v", result.Duration())
	}
}

func TestSubRange_AudioClip_MiddleTrim(t *testing.T) {
	a := NewAudioWithDuration("music.mp3", 60*time.Second)
	result := SubRange(a, 10*time.Second, 40*time.Second)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Duration() != 30*time.Second {
		t.Errorf("expected duration 30s, got %v", result.Duration())
	}
	if result.TrimStart() != 10*time.Second {
		t.Errorf("expected TrimStart 10s, got %v", result.TrimStart())
	}
	if result.TrimEnd() != 40*time.Second {
		t.Errorf("expected TrimEnd 40s, got %v", result.TrimEnd())
	}
}

func TestSubRange_AudioClip_AlreadyTrimmed(t *testing.T) {
	a := NewAudioWithDuration("music.mp3", 120*time.Second)
	trimmed := a.Trim(30*time.Second, 90*time.Second) // 60s clip, source 30-90

	result := SubRange(trimmed, 10*time.Second, 50*time.Second)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Duration() != 40*time.Second {
		t.Errorf("expected duration 40s, got %v", result.Duration())
	}
	if result.TrimStart() != 40*time.Second {
		t.Errorf("expected TrimStart 40s (source), got %v", result.TrimStart())
	}
	if result.TrimEnd() != 80*time.Second {
		t.Errorf("expected TrimEnd 80s (source), got %v", result.TrimEnd())
	}
}

func TestSubRange_ImageClip(t *testing.T) {
	img := NewImage("logo.png").WithDuration(10 * time.Second)
	result := SubRange(img, 3*time.Second, 8*time.Second)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Duration() != 5*time.Second {
		t.Errorf("expected duration 5s, got %v", result.Duration())
	}
}

func TestSubRange_ColorClip(t *testing.T) {
	c := NewColor("#FF0000", 1920, 1080).WithDuration(10 * time.Second)
	result := SubRange(c, 2*time.Second, 7*time.Second)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Duration() != 5*time.Second {
		t.Errorf("expected duration 5s, got %v", result.Duration())
	}
	// Verify color is preserved.
	cc, ok := result.(*ColorClip)
	if !ok {
		t.Fatal("expected *ColorClip")
	}
	if cc.Color != "#FF0000" {
		t.Errorf("expected color #FF0000, got %s", cc.Color)
	}
}

func TestSubRange_TextClip(t *testing.T) {
	txt := NewText("Hello", DefaultTextStyle()).WithDuration(8 * time.Second)
	result := SubRange(txt, 1*time.Second, 6*time.Second)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Duration() != 5*time.Second {
		t.Errorf("expected duration 5s, got %v", result.Duration())
	}
	tc, ok := result.(*TextClip)
	if !ok {
		t.Fatal("expected *TextClip")
	}
	if tc.Text != "Hello" {
		t.Errorf("expected text 'Hello', got %s", tc.Text)
	}
}

func TestSubRange_PreservesProperties(t *testing.T) {
	// Verify that SubRange preserves volume, fades, hasAudio, etc.
	v := NewVideoWithDuration("test.mp4", 30*time.Second)
	v = v.WithVolume(0.5).WithFadeIn(2 * time.Second)

	result := SubRange(v, 5*time.Second, 25*time.Second)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Volume() != 0.5 {
		t.Errorf("expected volume 0.5, got %f", result.Volume())
	}
	if result.FadeInDuration() != 2*time.Second {
		t.Errorf("expected fadeIn 2s, got %v", result.FadeInDuration())
	}
	if !result.HasAudio() {
		t.Error("expected HasAudio=true")
	}
	if !result.HasVideo() {
		t.Error("expected HasVideo=true")
	}
}

func TestSubRange_NegativeStart(t *testing.T) {
	// Negative start should be clamped to 0.
	v := NewVideoWithDuration("test.mp4", 30*time.Second)
	result := SubRange(v, -5*time.Second, 10*time.Second)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Duration() != 10*time.Second {
		t.Errorf("expected duration 10s, got %v", result.Duration())
	}
	if result.TrimStart() != 0 {
		t.Errorf("expected TrimStart 0, got %v", result.TrimStart())
	}
}
