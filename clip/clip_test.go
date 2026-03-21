package clip

import (
	"testing"
	"time"

	"github.com/ahmedhodiani/gomontage/effects"
)

func TestVideoClip_WithDuration(t *testing.T) {
	c := NewVideoWithDuration("test.mp4", 60*time.Second)

	if c.Duration() != 60*time.Second {
		t.Errorf("expected 60s, got %v", c.Duration())
	}
	if c.ClipType() != TypeVideo {
		t.Errorf("expected TypeVideo, got %v", c.ClipType())
	}
	if !c.HasVideo() {
		t.Error("expected HasVideo to be true")
	}
	if !c.HasAudio() {
		t.Error("expected HasAudio to be true")
	}
	if c.SourcePath() != "test.mp4" {
		t.Errorf("expected test.mp4, got %s", c.SourcePath())
	}
}

func TestVideoClip_Trim(t *testing.T) {
	c := NewVideoWithDuration("test.mp4", 60*time.Second)
	trimmed := c.Trim(10*time.Second, 30*time.Second)

	// Original should be unchanged.
	if c.Duration() != 60*time.Second {
		t.Errorf("original should be unchanged, got %v", c.Duration())
	}

	// Trimmed clip checks.
	if trimmed.Duration() != 20*time.Second {
		t.Errorf("expected 20s, got %v", trimmed.Duration())
	}
	if trimmed.TrimStart() != 10*time.Second {
		t.Errorf("expected trim start 10s, got %v", trimmed.TrimStart())
	}
	if trimmed.TrimEnd() != 30*time.Second {
		t.Errorf("expected trim end 30s, got %v", trimmed.TrimEnd())
	}
}

func TestVideoClip_Immutability(t *testing.T) {
	original := NewVideoWithDuration("test.mp4", 60*time.Second)

	_ = original.WithVolume(0.5)
	if original.Volume() != 1.0 {
		t.Error("WithVolume mutated original clip")
	}

	_ = original.WithFadeIn(1 * time.Second)
	if original.FadeInDuration() != 0 {
		t.Error("WithFadeIn mutated original clip")
	}

	_ = original.WithFadeOut(2 * time.Second)
	if original.FadeOutDuration() != 0 {
		t.Error("WithFadeOut mutated original clip")
	}
}

func TestVideoClip_AudioOnly(t *testing.T) {
	c := NewVideoWithDuration("test.mp4", 10*time.Second)
	audioOnly := c.AudioOnly()

	if audioOnly.HasVideo() {
		t.Error("AudioOnly should have HasVideo=false")
	}
	if !audioOnly.HasAudio() {
		t.Error("AudioOnly should have HasAudio=true")
	}
	// Original unchanged.
	if !c.HasVideo() {
		t.Error("original should still have video")
	}
}

func TestVideoClip_VideoOnly(t *testing.T) {
	c := NewVideoWithDuration("test.mp4", 10*time.Second)
	videoOnly := c.VideoOnly()

	if !videoOnly.HasVideo() {
		t.Error("VideoOnly should have HasVideo=true")
	}
	if videoOnly.HasAudio() {
		t.Error("VideoOnly should have HasAudio=false")
	}
}

func TestVideoClip_Chaining(t *testing.T) {
	c := NewVideoWithDuration("test.mp4", 60*time.Second)

	result := c.Trim(5*time.Second, 25*time.Second).
		WithVolume(0.5).
		WithFadeIn(1 * time.Second).
		WithFadeOut(2 * time.Second)

	if result.Duration() != 20*time.Second {
		t.Errorf("expected 20s, got %v", result.Duration())
	}
	if result.Volume() != 0.5 {
		t.Errorf("expected volume 0.5, got %f", result.Volume())
	}
	if result.FadeInDuration() != 1*time.Second {
		t.Errorf("expected fade in 1s, got %v", result.FadeInDuration())
	}
	if result.FadeOutDuration() != 2*time.Second {
		t.Errorf("expected fade out 2s, got %v", result.FadeOutDuration())
	}
}

func TestAudioClip(t *testing.T) {
	c := NewAudioWithDuration("narration.wav", 5*time.Minute)

	if c.ClipType() != TypeAudio {
		t.Errorf("expected TypeAudio, got %v", c.ClipType())
	}
	if c.HasVideo() {
		t.Error("AudioClip should not have video")
	}
	if !c.HasAudio() {
		t.Error("AudioClip should have audio")
	}
	if c.Duration() != 5*time.Minute {
		t.Errorf("expected 5m, got %v", c.Duration())
	}
}

func TestAudioClip_Transforms(t *testing.T) {
	c := NewAudioWithDuration("music.mp3", 3*time.Minute)

	quiet := c.WithVolume(0.2)
	if quiet.Volume() != 0.2 {
		t.Errorf("expected 0.2, got %f", quiet.Volume())
	}
	if c.Volume() != 1.0 {
		t.Error("original should be unchanged")
	}

	faded := c.WithFadeIn(3 * time.Second).WithFadeOut(5 * time.Second)
	if faded.FadeInDuration() != 3*time.Second {
		t.Errorf("expected 3s, got %v", faded.FadeInDuration())
	}
	if faded.FadeOutDuration() != 5*time.Second {
		t.Errorf("expected 5s, got %v", faded.FadeOutDuration())
	}

	trimmed := c.Trim(30*time.Second, 90*time.Second)
	if trimmed.Duration() != 60*time.Second {
		t.Errorf("expected 60s, got %v", trimmed.Duration())
	}
}

func TestImageClip(t *testing.T) {
	c := NewImage("logo.png")

	if c.ClipType() != TypeImage {
		t.Errorf("expected TypeImage, got %v", c.ClipType())
	}
	if c.Duration() != 5*time.Second {
		t.Errorf("expected default 5s, got %v", c.Duration())
	}
	if !c.HasVideo() {
		t.Error("ImageClip should have video")
	}

	custom := c.WithDuration(10 * time.Second)
	if custom.Duration() != 10*time.Second {
		t.Errorf("expected 10s, got %v", custom.Duration())
	}
	if c.Duration() != 5*time.Second {
		t.Error("original should be unchanged")
	}
}

func TestTextClip(t *testing.T) {
	c := NewText("Hello World", TextStyle{
		Size:  72,
		Color: "#FF0000",
	})

	if c.ClipType() != TypeText {
		t.Errorf("expected TypeText, got %v", c.ClipType())
	}
	if c.Text != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", c.Text)
	}
	if c.Style.Size != 72 {
		t.Errorf("expected size 72, got %d", c.Style.Size)
	}
	if c.Duration() != 5*time.Second {
		t.Errorf("expected default 5s, got %v", c.Duration())
	}

	custom := c.WithDuration(3 * time.Second)
	if custom.Duration() != 3*time.Second {
		t.Errorf("expected 3s, got %v", custom.Duration())
	}
	if custom.Text != "Hello World" {
		t.Error("text should be preserved")
	}
}

func TestTextClip_DefaultStyle(t *testing.T) {
	c := NewText("Test", TextStyle{})

	if c.Style.Size != 48 {
		t.Errorf("expected default size 48, got %d", c.Style.Size)
	}
	if c.Style.Color != "#FFFFFF" {
		t.Errorf("expected default color #FFFFFF, got %s", c.Style.Color)
	}
}

func TestColorClip(t *testing.T) {
	c := NewColor("#000000", 1920, 1080)

	if c.ClipType() != TypeColor {
		t.Errorf("expected TypeColor, got %v", c.ClipType())
	}
	if c.Color != "#000000" {
		t.Errorf("expected #000000, got %s", c.Color)
	}
	if c.Width() != 1920 || c.Height() != 1080 {
		t.Errorf("expected 1920x1080, got %dx%d", c.Width(), c.Height())
	}

	short := c.WithDuration(2 * time.Second)
	if short.Duration() != 2*time.Second {
		t.Errorf("expected 2s, got %v", short.Duration())
	}
	if short.Color != "#000000" {
		t.Error("color should be preserved")
	}
}

func TestClipInterface(t *testing.T) {
	// Verify all types implement the Clip interface.
	var clips []Clip
	clips = append(clips, NewVideoWithDuration("v.mp4", time.Second))
	clips = append(clips, NewAudioWithDuration("a.wav", time.Second))
	clips = append(clips, NewImage("i.png"))
	clips = append(clips, NewText("text", TextStyle{}))
	clips = append(clips, NewColor("#FFF", 100, 100))

	expectedTypes := []Type{TypeVideo, TypeAudio, TypeImage, TypeText, TypeColor}
	for i, c := range clips {
		if c.ClipType() != expectedTypes[i] {
			t.Errorf("clip %d: expected type %v, got %v", i, expectedTypes[i], c.ClipType())
		}
	}
}

func TestType_String(t *testing.T) {
	tests := []struct {
		t    Type
		want string
	}{
		{TypeVideo, "video"},
		{TypeAudio, "audio"},
		{TypeImage, "image"},
		{TypeText, "text"},
		{TypeColor, "color"},
		{Type(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.t.String(); got != tt.want {
			t.Errorf("Type(%d).String() = %q, want %q", int(tt.t), got, tt.want)
		}
	}
}

func TestVideoClip_WithEffect(t *testing.T) {
	original := NewVideoWithDuration("test.mp4", 30*time.Second)

	// Apply a single effect.
	withSpeed := original.WithEffect(effects.SpeedUp(2.0))
	if len(withSpeed.Effects()) != 1 {
		t.Fatalf("expected 1 effect, got %d", len(withSpeed.Effects()))
	}
	if withSpeed.Effects()[0].Name() != "speed" {
		t.Errorf("expected effect name 'speed', got %q", withSpeed.Effects()[0].Name())
	}

	// Original should be unchanged.
	if len(original.Effects()) != 0 {
		t.Error("WithEffect mutated original clip")
	}

	// Stack multiple effects.
	stacked := original.
		WithEffect(effects.FadeIn(1 * time.Second)).
		WithEffect(effects.FadeOut(2 * time.Second)).
		WithEffect(effects.SpeedUp(1.5))
	if len(stacked.Effects()) != 3 {
		t.Fatalf("expected 3 effects, got %d", len(stacked.Effects()))
	}
	if stacked.Effects()[0].Name() != "fade_in" {
		t.Errorf("expected first effect 'fade_in', got %q", stacked.Effects()[0].Name())
	}
	if stacked.Effects()[1].Name() != "fade_out" {
		t.Errorf("expected second effect 'fade_out', got %q", stacked.Effects()[1].Name())
	}
	if stacked.Effects()[2].Name() != "speed" {
		t.Errorf("expected third effect 'speed', got %q", stacked.Effects()[2].Name())
	}
}

func TestVideoClip_WithEffect_Chaining(t *testing.T) {
	c := NewVideoWithDuration("test.mp4", 60*time.Second)

	result := c.Trim(5*time.Second, 25*time.Second).
		WithVolume(0.5).
		WithEffect(effects.FadeIn(1 * time.Second)).
		WithEffect(effects.SpeedUp(2.0))

	if result.Duration() != 20*time.Second {
		t.Errorf("expected 20s, got %v", result.Duration())
	}
	if result.Volume() != 0.5 {
		t.Errorf("expected volume 0.5, got %f", result.Volume())
	}
	if len(result.Effects()) != 2 {
		t.Fatalf("expected 2 effects, got %d", len(result.Effects()))
	}
}

func TestAudioClip_WithEffect(t *testing.T) {
	original := NewAudioWithDuration("music.mp3", 3*time.Minute)

	withVol := original.WithEffect(effects.Volume(0.5))
	if len(withVol.Effects()) != 1 {
		t.Fatalf("expected 1 effect, got %d", len(withVol.Effects()))
	}
	if withVol.Effects()[0].Name() != "volume" {
		t.Errorf("expected effect name 'volume', got %q", withVol.Effects()[0].Name())
	}

	// Original should be unchanged.
	if len(original.Effects()) != 0 {
		t.Error("WithEffect mutated original clip")
	}

	// Stack effects.
	stacked := original.
		WithEffect(effects.AudioFadeIn(2 * time.Second)).
		WithEffect(effects.Normalize())
	if len(stacked.Effects()) != 2 {
		t.Fatalf("expected 2 effects, got %d", len(stacked.Effects()))
	}
}

func TestImageClip_WithEffect(t *testing.T) {
	original := NewImage("logo.png")

	withFade := original.WithEffect(effects.FadeIn(1 * time.Second))
	if len(withFade.Effects()) != 1 {
		t.Fatalf("expected 1 effect, got %d", len(withFade.Effects()))
	}
	if withFade.Effects()[0].Name() != "fade_in" {
		t.Errorf("expected effect name 'fade_in', got %q", withFade.Effects()[0].Name())
	}

	// Original unchanged.
	if len(original.Effects()) != 0 {
		t.Error("WithEffect mutated original clip")
	}
}

func TestTextClip_WithEffect(t *testing.T) {
	original := NewText("Title", DefaultTextStyle())

	withFade := original.WithEffect(effects.FadeIn(1 * time.Second))
	if len(withFade.Effects()) != 1 {
		t.Fatalf("expected 1 effect, got %d", len(withFade.Effects()))
	}
	if withFade.Text != "Title" {
		t.Error("text should be preserved")
	}

	// Original unchanged.
	if len(original.Effects()) != 0 {
		t.Error("WithEffect mutated original clip")
	}
}

func TestColorClip_WithEffect(t *testing.T) {
	original := NewColor("#FF0000", 1920, 1080)

	withFade := original.WithEffect(effects.FadeOut(2 * time.Second))
	if len(withFade.Effects()) != 1 {
		t.Fatalf("expected 1 effect, got %d", len(withFade.Effects()))
	}
	if withFade.Color != "#FF0000" {
		t.Error("color should be preserved")
	}

	// Original unchanged.
	if len(original.Effects()) != 0 {
		t.Error("WithEffect mutated original clip")
	}
}

func TestEffects_Immutability(t *testing.T) {
	// Verify that the returned effects slice is a copy, not the internal slice.
	c := NewVideoWithDuration("test.mp4", 10*time.Second).
		WithEffect(effects.FadeIn(1 * time.Second))

	effs := c.Effects()
	effs[0] = effects.SpeedUp(2.0) // Mutate the returned slice.

	// Internal effects should not be affected.
	if c.Effects()[0].Name() != "fade_in" {
		t.Error("Effects() returned the internal slice instead of a copy")
	}
}
