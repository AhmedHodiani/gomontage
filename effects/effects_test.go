package effects

import (
	"testing"
	"time"
)

func TestFadeIn(t *testing.T) {
	e := FadeIn(1 * time.Second)
	if e.Name() != "fade_in" {
		t.Errorf("expected name 'fade_in', got %q", e.Name())
	}
	if e.Target() != TargetVideo {
		t.Errorf("expected TargetVideo, got %v", e.Target())
	}
	if e.FilterName() != "fade" {
		t.Errorf("expected 'fade', got %q", e.FilterName())
	}
	params := e.FilterParams()
	if params["t"] != "in" {
		t.Errorf("expected t=in, got %q", params["t"])
	}
	if params["d"] != "1" {
		t.Errorf("expected d=1, got %q", params["d"])
	}
}

func TestFadeOut(t *testing.T) {
	e := FadeOut(2 * time.Second)
	if e.Name() != "fade_out" {
		t.Errorf("expected name 'fade_out', got %q", e.Name())
	}
	if e.FilterName() != "fade" {
		t.Errorf("expected 'fade', got %q", e.FilterName())
	}
	params := e.FilterParams()
	if params["t"] != "out" {
		t.Errorf("expected t=out, got %q", params["t"])
	}
}

func TestFadeOutAt(t *testing.T) {
	e := FadeOutAt(10*time.Second, 2*time.Second)
	params := e.FilterParams()
	if params["st"] != "10" {
		t.Errorf("expected st=10, got %q", params["st"])
	}
	if params["d"] != "2" {
		t.Errorf("expected d=2, got %q", params["d"])
	}
}

func TestSpeedUp(t *testing.T) {
	e := SpeedUp(2.0)
	if e.Name() != "speed" {
		t.Errorf("expected name 'speed', got %q", e.Name())
	}
	if e.Target() != TargetVideo {
		t.Errorf("expected TargetVideo, got %v", e.Target())
	}
	if e.FilterName() != "setpts" {
		t.Errorf("expected 'setpts', got %q", e.FilterName())
	}
	if e.Factor() != 2.0 {
		t.Errorf("expected factor 2.0, got %f", e.Factor())
	}
	params := e.FilterParams()
	if params["expr"] != "0.5*PTS" {
		t.Errorf("expected 0.5*PTS, got %q", params["expr"])
	}
}

func TestSlowDown(t *testing.T) {
	e := SlowDown(2.0)
	if e.Factor() != 0.5 {
		t.Errorf("expected factor 0.5, got %f", e.Factor())
	}
	params := e.FilterParams()
	if params["expr"] != "2*PTS" {
		t.Errorf("expected 2*PTS, got %q", params["expr"])
	}
}

func TestSlowDown_Zero(t *testing.T) {
	e := SlowDown(0)
	if e.Factor() != 1.0 {
		t.Errorf("expected factor 1.0 for zero input, got %f", e.Factor())
	}
}

func TestAudioFadeIn(t *testing.T) {
	e := AudioFadeIn(3 * time.Second)
	if e.Name() != "audio_fade_in" {
		t.Errorf("expected 'audio_fade_in', got %q", e.Name())
	}
	if e.Target() != TargetAudio {
		t.Errorf("expected TargetAudio, got %v", e.Target())
	}
	if e.FilterName() != "afade" {
		t.Errorf("expected 'afade', got %q", e.FilterName())
	}
	if e.Dur() != 3*time.Second {
		t.Errorf("expected 3s, got %v", e.Dur())
	}
}

func TestAudioFadeOut(t *testing.T) {
	e := AudioFadeOut(2 * time.Second)
	if e.FilterName() != "afade" {
		t.Errorf("expected 'afade', got %q", e.FilterName())
	}
	params := e.FilterParams()
	if params["t"] != "out" {
		t.Errorf("expected t=out, got %q", params["t"])
	}
}

func TestVolume(t *testing.T) {
	e := Volume(0.5)
	if e.Name() != "volume" {
		t.Errorf("expected 'volume', got %q", e.Name())
	}
	if e.Target() != TargetAudio {
		t.Errorf("expected TargetAudio, got %v", e.Target())
	}
	if e.Level() != 0.5 {
		t.Errorf("expected 0.5, got %f", e.Level())
	}
	params := e.FilterParams()
	if params["volume"] != "0.5" {
		t.Errorf("expected volume=0.5, got %q", params["volume"])
	}
}

func TestNormalize(t *testing.T) {
	e := Normalize()
	if e.Name() != "normalize" {
		t.Errorf("expected 'normalize', got %q", e.Name())
	}
	if e.FilterName() != "loudnorm" {
		t.Errorf("expected 'loudnorm', got %q", e.FilterName())
	}
	if e.TargetLUFS() != -16.0 {
		t.Errorf("expected -16.0, got %f", e.TargetLUFS())
	}
}

func TestNormalizeTo(t *testing.T) {
	e := NormalizeTo(-23.0)
	if e.TargetLUFS() != -23.0 {
		t.Errorf("expected -23.0, got %f", e.TargetLUFS())
	}
}

func TestAudioSpeed(t *testing.T) {
	e := AudioSpeed(1.5)
	if e.Name() != "audio_speed" {
		t.Errorf("expected 'audio_speed', got %q", e.Name())
	}
	if e.FilterName() != "atempo" {
		t.Errorf("expected 'atempo', got %q", e.FilterName())
	}
	if e.Factor() != 1.5 {
		t.Errorf("expected 1.5, got %f", e.Factor())
	}
}

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{1.0, "1"},
		{0.5, "0.5"},
		{2.0, "2"},
		{1.5, "1.5"},
		{0.123456, "0.123456"},
	}

	for _, tt := range tests {
		got := formatFloat(tt.input)
		if got != tt.want {
			t.Errorf("formatFloat(%f) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestEffectInterface(t *testing.T) {
	// Verify all types implement Effect.
	effs := []Effect{
		FadeIn(time.Second),
		FadeOut(time.Second),
		SpeedUp(2.0),
		AudioFadeIn(time.Second),
		AudioFadeOut(time.Second),
		Volume(0.5),
		Normalize(),
		AudioSpeed(1.5),
	}

	for i, e := range effs {
		if e.Name() == "" {
			t.Errorf("effect %d has empty name", i)
		}
		if e.FilterName() == "" {
			t.Errorf("effect %d has empty filter name", i)
		}
		if e.FilterParams() == nil {
			t.Errorf("effect %d has nil params", i)
		}
	}
}

func TestDurationFactor_NoChange(t *testing.T) {
	// Effects that don't change duration should return 1.0.
	noChange := []Effect{
		FadeIn(time.Second),
		FadeOut(time.Second),
		AudioFadeIn(time.Second),
		AudioFadeOut(time.Second),
		Volume(0.5),
		Normalize(),
	}

	for _, e := range noChange {
		if f := e.DurationFactor(); f != 1.0 {
			t.Errorf("%s.DurationFactor() = %f, want 1.0", e.Name(), f)
		}
	}
}

func TestDurationFactor_Speed(t *testing.T) {
	tests := []struct {
		name   string
		effect Effect
		want   float64
	}{
		{"SpeedUp 2x", SpeedUp(2.0), 0.5},
		{"SpeedUp 0.5x", SpeedUp(0.5), 2.0},
		{"SlowDown 2x", SlowDown(2.0), 2.0},
		{"SlowDown 0 (safe)", SlowDown(0), 1.0},
		{"AudioSpeed 1.5x", AudioSpeed(1.5), 1.0 / 1.5},
		{"AudioSpeed 0.5x", AudioSpeed(0.5), 2.0},
	}

	for _, tt := range tests {
		got := tt.effect.DurationFactor()
		if got != tt.want {
			t.Errorf("%s: DurationFactor() = %f, want %f", tt.name, got, tt.want)
		}
	}
}
