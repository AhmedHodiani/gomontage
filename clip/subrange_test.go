package clip

import (
	"testing"
	"time"

	"github.com/ahmedhodiani/gomontage/effects"
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

// --- Speed effect edge cases (ratio-based source-time translation) ---

func TestSubRange_VideoClip_SpeedUp(t *testing.T) {
	// 30s source with SpeedUp(2.0): Duration()=15s, TrimStart=0, TrimEnd=30s.
	// sourceWindow=30s, localDur=15s, ratio=2.0.
	// SubRange(5s, 10s) → source: 0 + 5*2=10s, 0 + 10*2=20s.
	v := NewVideoWithDuration("test.mp4", 30*time.Second).
		WithEffect(effects.SpeedUp(2.0))

	if v.Duration() != 15*time.Second {
		t.Fatalf("precondition: expected duration 15s after SpeedUp(2.0), got %v", v.Duration())
	}

	result := SubRange(v, 5*time.Second, 10*time.Second)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Duration() != 5*time.Second {
		t.Errorf("expected duration 5s, got %v", result.Duration())
	}
	if result.TrimStart() != 10*time.Second {
		t.Errorf("expected TrimStart 10s (source), got %v", result.TrimStart())
	}
	if result.TrimEnd() != 20*time.Second {
		t.Errorf("expected TrimEnd 20s (source), got %v", result.TrimEnd())
	}
}

func TestSubRange_VideoClip_SlowDown(t *testing.T) {
	// 30s source with SlowDown(2.0) = SpeedUp(0.5): Duration()=60s, ratio=0.5.
	// SubRange(10s, 30s) → source: 0 + 10*0.5=5s, 0 + 30*0.5=15s.
	v := NewVideoWithDuration("test.mp4", 30*time.Second).
		WithEffect(effects.SlowDown(2.0))

	if v.Duration() != 60*time.Second {
		t.Fatalf("precondition: expected duration 60s after SlowDown(2.0), got %v", v.Duration())
	}

	result := SubRange(v, 10*time.Second, 30*time.Second)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Duration() != 20*time.Second {
		t.Errorf("expected duration 20s, got %v", result.Duration())
	}
	if result.TrimStart() != 5*time.Second {
		t.Errorf("expected TrimStart 5s (source), got %v", result.TrimStart())
	}
	if result.TrimEnd() != 15*time.Second {
		t.Errorf("expected TrimEnd 15s (source), got %v", result.TrimEnd())
	}
}

func TestSubRange_VideoClip_TrimThenSpeedUp(t *testing.T) {
	// Compound: Trim(20s, 40s) then SpeedUp(2.0).
	// Source window = 20s (20-40), localDur = 10s (speed halves it), ratio = 2.0.
	// SubRange(2s, 8s) → source: 20 + 2*2=24s, 20 + 8*2=36s.
	v := NewVideoWithDuration("test.mp4", 60*time.Second).
		Trim(20*time.Second, 40*time.Second).
		WithEffect(effects.SpeedUp(2.0))

	if v.Duration() != 10*time.Second {
		t.Fatalf("precondition: expected duration 10s, got %v", v.Duration())
	}

	result := SubRange(v, 2*time.Second, 8*time.Second)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Duration() != 6*time.Second {
		t.Errorf("expected duration 6s, got %v", result.Duration())
	}
	if result.TrimStart() != 24*time.Second {
		t.Errorf("expected TrimStart 24s (source), got %v", result.TrimStart())
	}
	if result.TrimEnd() != 36*time.Second {
		t.Errorf("expected TrimEnd 36s (source), got %v", result.TrimEnd())
	}
}

func TestSubRange_AudioClip_WithAudioSpeed(t *testing.T) {
	// AudioClip (no video) with AudioSpeed(2.0): Duration halved.
	// 60s source → Duration()=30s, ratio=2.0.
	// SubRange(5s, 25s) → source: 0 + 5*2=10s, 0 + 25*2=50s.
	a := NewAudioWithDuration("music.mp3", 60*time.Second).
		WithEffect(effects.AudioSpeed(2.0))

	if a.Duration() != 30*time.Second {
		t.Fatalf("precondition: expected duration 30s after AudioSpeed(2.0), got %v", a.Duration())
	}

	result := SubRange(a, 5*time.Second, 25*time.Second)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Duration() != 20*time.Second {
		t.Errorf("expected duration 20s, got %v", result.Duration())
	}
	if result.TrimStart() != 10*time.Second {
		t.Errorf("expected TrimStart 10s (source), got %v", result.TrimStart())
	}
	if result.TrimEnd() != 50*time.Second {
		t.Errorf("expected TrimEnd 50s (source), got %v", result.TrimEnd())
	}
}

func TestSubRange_AudioClip_TrimThenAudioSpeed(t *testing.T) {
	// Compound: Trim(30s, 90s) then AudioSpeed(3.0).
	// Source window = 60s, localDur = 20s, ratio = 3.0.
	// SubRange(5s, 15s) → source: 30 + 5*3=45s, 30 + 15*3=75s.
	a := NewAudioWithDuration("music.mp3", 120*time.Second).
		Trim(30*time.Second, 90*time.Second).
		WithEffect(effects.AudioSpeed(3.0))

	if a.Duration() != 20*time.Second {
		t.Fatalf("precondition: expected duration 20s, got %v", a.Duration())
	}

	result := SubRange(a, 5*time.Second, 15*time.Second)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Duration() != 10*time.Second {
		t.Errorf("expected duration 10s, got %v", result.Duration())
	}
	if result.TrimStart() != 45*time.Second {
		t.Errorf("expected TrimStart 45s (source), got %v", result.TrimStart())
	}
	if result.TrimEnd() != 75*time.Second {
		t.Errorf("expected TrimEnd 75s (source), got %v", result.TrimEnd())
	}
}

// --- Property preservation edge cases ---

func TestSubRange_PreservesFadeOut(t *testing.T) {
	v := NewVideoWithDuration("test.mp4", 30*time.Second).
		WithFadeOut(3 * time.Second)

	result := SubRange(v, 5*time.Second, 25*time.Second)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.FadeOutDuration() != 3*time.Second {
		t.Errorf("expected fadeOut 3s preserved, got %v", result.FadeOutDuration())
	}
}

func TestSubRange_PreservesVideoOnly(t *testing.T) {
	v := NewVideoWithDuration("test.mp4", 30*time.Second).VideoOnly()

	result := SubRange(v, 0, 20*time.Second)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.HasAudio() {
		t.Error("expected HasAudio=false (VideoOnly) to be preserved")
	}
	if !result.HasVideo() {
		t.Error("expected HasVideo=true")
	}
}

func TestSubRange_PreservesAudioOnly(t *testing.T) {
	v := NewVideoWithDuration("test.mp4", 30*time.Second).AudioOnly()

	result := SubRange(v, 0, 20*time.Second)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.HasAudio() {
		t.Error("expected HasAudio=true")
	}
	if result.HasVideo() {
		t.Error("expected HasVideo=false (AudioOnly) to be preserved")
	}
}

func TestSubRange_PreservesPosition(t *testing.T) {
	pos := Position{X: 100, Y: 200}
	v := NewVideoWithDuration("test.mp4", 30*time.Second).
		WithPosition(pos)

	result := SubRange(v, 5*time.Second, 25*time.Second)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Pos() != pos {
		t.Errorf("expected position %+v, got %+v", pos, result.Pos())
	}
}

func TestSubRange_PreservesSize(t *testing.T) {
	v := NewVideoWithDuration("test.mp4", 30*time.Second).
		WithSize(640, 480)

	result := SubRange(v, 5*time.Second, 25*time.Second)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Width() != 640 {
		t.Errorf("expected width 640, got %d", result.Width())
	}
	if result.Height() != 480 {
		t.Errorf("expected height 480, got %d", result.Height())
	}
}

func TestSubRange_PreservesEffectsList(t *testing.T) {
	v := NewVideoWithDuration("test.mp4", 30*time.Second).
		WithEffect(effects.SpeedUp(2.0)).
		WithEffect(effects.FadeIn(1 * time.Second))

	result := SubRange(v, 0, v.Duration())
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// Full window returns the same clip, so effects are trivially preserved.
	// Test a partial window too.
	partial := SubRange(v, 2*time.Second, 10*time.Second)
	if partial == nil {
		t.Fatal("expected non-nil partial result")
	}
	effs := partial.Effects()
	if len(effs) != 2 {
		t.Fatalf("expected 2 effects preserved, got %d", len(effs))
	}
	if effs[0].Name() != "speed" {
		t.Errorf("expected first effect 'speed', got %q", effs[0].Name())
	}
	if effs[1].Name() != "fade_in" {
		t.Errorf("expected second effect 'fade_in', got %q", effs[1].Name())
	}
}

func TestSubRange_AudioClip_PreservesProperties(t *testing.T) {
	a := NewAudioWithDuration("music.mp3", 60*time.Second).
		WithVolume(0.3).
		WithFadeIn(2 * time.Second).
		WithFadeOut(5 * time.Second)

	result := SubRange(a, 10*time.Second, 50*time.Second)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Volume() != 0.3 {
		t.Errorf("expected volume 0.3, got %f", result.Volume())
	}
	if result.FadeInDuration() != 2*time.Second {
		t.Errorf("expected fadeIn 2s, got %v", result.FadeInDuration())
	}
	if result.FadeOutDuration() != 5*time.Second {
		t.Errorf("expected fadeOut 5s, got %v", result.FadeOutDuration())
	}
}

func TestSubRange_ImageClip_PreservesProperties(t *testing.T) {
	img := NewImage("logo.png").
		WithDuration(10*time.Second).
		WithSize(800, 600).
		WithPosition(Position{X: 50, Y: 50}).
		WithFadeIn(1 * time.Second).
		WithFadeOut(1 * time.Second)

	result := SubRange(img, 2*time.Second, 8*time.Second)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Width() != 800 {
		t.Errorf("expected width 800, got %d", result.Width())
	}
	if result.Height() != 600 {
		t.Errorf("expected height 600, got %d", result.Height())
	}
	if result.Pos() != (Position{X: 50, Y: 50}) {
		t.Errorf("expected position {50, 50}, got %+v", result.Pos())
	}
	if result.FadeInDuration() != 1*time.Second {
		t.Errorf("expected fadeIn 1s, got %v", result.FadeInDuration())
	}
	if result.FadeOutDuration() != 1*time.Second {
		t.Errorf("expected fadeOut 1s, got %v", result.FadeOutDuration())
	}
}

func TestSubRange_ColorClip_PreservesDimensions(t *testing.T) {
	c := NewColor("#00FF00", 1920, 1080).WithDuration(10 * time.Second)
	result := SubRange(c, 1*time.Second, 9*time.Second)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	cc := result.(*ColorClip)
	if cc.Width() != 1920 {
		t.Errorf("expected width 1920, got %d", cc.Width())
	}
	if cc.Height() != 1080 {
		t.Errorf("expected height 1080, got %d", cc.Height())
	}
}

func TestSubRange_TextClip_PreservesStyle(t *testing.T) {
	style := TextStyle{
		Font:        "/fonts/bold.ttf",
		Size:        72,
		Color:       "#FF0000",
		BorderWidth: 3,
		BorderColor: "#000000",
		Position:    Position{X: 100, Y: 200},
	}
	txt := NewText("Title", style).WithDuration(8 * time.Second)

	result := SubRange(txt, 1*time.Second, 7*time.Second)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	tc := result.(*TextClip)
	if tc.Style.Font != "/fonts/bold.ttf" {
		t.Errorf("expected font preserved, got %q", tc.Style.Font)
	}
	if tc.Style.Size != 72 {
		t.Errorf("expected size 72, got %d", tc.Style.Size)
	}
	if tc.Style.Color != "#FF0000" {
		t.Errorf("expected color #FF0000, got %s", tc.Style.Color)
	}
	if tc.Style.BorderWidth != 3 {
		t.Errorf("expected border width 3, got %d", tc.Style.BorderWidth)
	}
}

// --- Boundary / degenerate input edge cases ---

func TestSubRange_ZeroDurationClip(t *testing.T) {
	// A clip with Duration()=0 should always return nil from SubRange.
	v := NewVideoWithDuration("test.mp4", 0)

	result := SubRange(v, 0, 0)
	if result != nil {
		t.Error("expected nil for zero-duration clip with zero window")
	}

	result = SubRange(v, 0, 5*time.Second)
	if result != nil {
		t.Error("expected nil for zero-duration clip (end clamped to 0)")
	}
}

func TestSubRange_FullWindowViaClamping(t *testing.T) {
	// Both bounds clamped → should produce full window and return same clip.
	v := NewVideoWithDuration("test.mp4", 30*time.Second)
	result := SubRange(v, -10*time.Second, 100*time.Second)

	if result != v {
		t.Error("SubRange with both bounds clamped to full window should return original clip")
	}
}

func TestSubRange_VerySmallWindow(t *testing.T) {
	v := NewVideoWithDuration("test.mp4", 30*time.Second)
	result := SubRange(v, 10*time.Second, 10*time.Second+time.Millisecond)

	if result == nil {
		t.Fatal("expected non-nil result for 1ms window")
	}
	if result.Duration() != time.Millisecond {
		t.Errorf("expected duration 1ms, got %v", result.Duration())
	}
}

func TestSubRange_VerySmallWindowWithSpeed(t *testing.T) {
	// 30s source, SpeedUp(2) → 15s local. 1ms window at 7s.
	// ratio=2.0, source: 0 + 7000ms*2 = 14000ms, 0 + 7001ms*2 = 14002ms.
	v := NewVideoWithDuration("test.mp4", 30*time.Second).
		WithEffect(effects.SpeedUp(2.0))

	result := SubRange(v, 7*time.Second, 7*time.Second+time.Millisecond)
	if result == nil {
		t.Fatal("expected non-nil result for 1ms window with speed")
	}
	if result.Duration() != time.Millisecond {
		t.Errorf("expected duration 1ms, got %v", result.Duration())
	}
	// Source coordinates should be 14s and 14.002s.
	if result.TrimStart() != 14*time.Second {
		t.Errorf("expected TrimStart 14s, got %v", result.TrimStart())
	}
	expectedEnd := 14*time.Second + 2*time.Millisecond
	if result.TrimEnd() != expectedEnd {
		t.Errorf("expected TrimEnd %v, got %v", expectedEnd, result.TrimEnd())
	}
}

// --- Generated clip trim metadata edge cases ---

func TestSubRange_ImageClip_TrimMetadata(t *testing.T) {
	img := NewImage("bg.png").WithDuration(10 * time.Second)
	result := SubRange(img, 2*time.Second, 7*time.Second)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// For generated clips, TrimEnd should equal the new duration.
	if result.TrimEnd() != 5*time.Second {
		t.Errorf("expected TrimEnd 5s, got %v", result.TrimEnd())
	}
	// TrimStart should remain 0 (images don't have source offsets).
	if result.TrimStart() != 0 {
		t.Errorf("expected TrimStart 0, got %v", result.TrimStart())
	}
}

func TestSubRange_ColorClip_TrimMetadata(t *testing.T) {
	c := NewColor("#000000", 1920, 1080).WithDuration(10 * time.Second)
	result := SubRange(c, 3*time.Second, 8*time.Second)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.TrimEnd() != 5*time.Second {
		t.Errorf("expected TrimEnd 5s, got %v", result.TrimEnd())
	}
	if result.TrimStart() != 0 {
		t.Errorf("expected TrimStart 0, got %v", result.TrimStart())
	}
}

func TestSubRange_TextClip_TrimMetadata(t *testing.T) {
	txt := NewText("Hello", DefaultTextStyle()).WithDuration(8 * time.Second)
	result := SubRange(txt, 1*time.Second, 6*time.Second)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.TrimEnd() != 5*time.Second {
		t.Errorf("expected TrimEnd 5s, got %v", result.TrimEnd())
	}
	if result.TrimStart() != 0 {
		t.Errorf("expected TrimStart 0, got %v", result.TrimStart())
	}
}
