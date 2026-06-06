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

func TestFadeIn_Dur(t *testing.T) {
	e := FadeIn(500 * time.Millisecond)
	if e.Dur() != 500*time.Millisecond {
		t.Errorf("expected 500ms, got %v", e.Dur())
	}
}

func TestFadeIn_Params(t *testing.T) {
	e := FadeIn(1500 * time.Millisecond)
	p := e.FilterParams()
	if p["st"] != "0" {
		t.Errorf("expected st=0, got %q", p["st"])
	}
	if p["d"] != "1.5" {
		t.Errorf("expected d=1.5, got %q", p["d"])
	}
}

func TestFadeOut_StartAt(t *testing.T) {
	e := FadeOut(3 * time.Second)
	if e.StartAt() != 0 {
		t.Errorf("expected StartAt()=0 for default fade-out, got %v", e.StartAt())
	}
}

func TestFadeOut_Dur(t *testing.T) {
	e := FadeOut(3 * time.Second)
	if e.Dur() != 3*time.Second {
		t.Errorf("expected 3s, got %v", e.Dur())
	}
}

func TestFadeOut_Target(t *testing.T) {
	e := FadeOut(time.Second)
	if e.Target() != TargetVideo {
		t.Errorf("expected TargetVideo, got %v", e.Target())
	}
}

func TestFadeOut_Params(t *testing.T) {
	e := FadeOut(500 * time.Millisecond)
	p := e.FilterParams()
	if p["st"] != "0" {
		t.Errorf("expected st=0 for default fade-out, got %q", p["st"])
	}
	if p["d"] != "0.5" {
		t.Errorf("expected d=0.5, got %q", p["d"])
	}
}

func TestFadeOutAt_Full(t *testing.T) {
	e := FadeOutAt(5*time.Second, 750*time.Millisecond)
	if e.Name() != "fade_out" {
		t.Errorf("expected 'fade_out', got %q", e.Name())
	}
	if e.Target() != TargetVideo {
		t.Errorf("expected TargetVideo, got %v", e.Target())
	}
	if e.FilterName() != "fade" {
		t.Errorf("expected 'fade', got %q", e.FilterName())
	}
	if e.Dur() != 750*time.Millisecond {
		t.Errorf("expected 750ms, got %v", e.Dur())
	}
	if e.StartAt() != 5*time.Second {
		t.Errorf("expected 5s, got %v", e.StartAt())
	}
	p := e.FilterParams()
	if p["t"] != "out" {
		t.Errorf("expected t=out, got %q", p["t"])
	}
	if p["st"] != "5" {
		t.Errorf("expected st=5, got %q", p["st"])
	}
	if p["d"] != "0.75" {
		t.Errorf("expected d=0.75, got %q", p["d"])
	}
}

func TestSpeedUp_Zero(t *testing.T) {
	e := SpeedUp(0)
	if e.Factor() != 0 {
		t.Errorf("expected factor 0, got %f", e.Factor())
	}
	p := e.FilterParams()
	if p["expr"] != "+Inf*PTS" {
		t.Errorf("expected +Inf*PTS for factor 0, got %q", p["expr"])
	}
	if f := e.DurationFactor(); f != 1.0 {
		t.Errorf("expected DurationFactor=1.0 for SpeedUp(0) (split by zero yields +Inf, but safety returns 1.0), got %f", f)
	}
}

func TestSpeedUp_One(t *testing.T) {
	e := SpeedUp(1.0)
	p := e.FilterParams()
	if p["expr"] != "1*PTS" {
		t.Errorf("expected 1*PTS, got %q", p["expr"])
	}
	if f := e.DurationFactor(); f != 1.0 {
		t.Errorf("expected DurationFactor=1.0 for SpeedUp(1.0), got %f", f)
	}
}

func TestSlowDown_NameTarget(t *testing.T) {
	e := SlowDown(3.0)
	if e.Name() != "speed" {
		t.Errorf("expected 'speed', got %q", e.Name())
	}
	if e.Target() != TargetVideo {
		t.Errorf("expected TargetVideo, got %v", e.Target())
	}
	if e.FilterName() != "setpts" {
		t.Errorf("expected 'setpts', got %q", e.FilterName())
	}
}

func TestSlowDown_One(t *testing.T) {
	e := SlowDown(1.0)
	if e.Factor() != 1.0 {
		t.Errorf("expected factor 1.0, got %f", e.Factor())
	}
	p := e.FilterParams()
	if p["expr"] != "1*PTS" {
		t.Errorf("expected 1*PTS, got %q", p["expr"])
	}
}

func TestSlowDown_Negative(t *testing.T) {
	e := SlowDown(-2.0)
	if e.Factor() != -0.5 {
		t.Errorf("expected factor -0.5, got %f", e.Factor())
	}
}

func TestAudioFadeIn_Params(t *testing.T) {
	e := AudioFadeIn(2500 * time.Millisecond)
	p := e.FilterParams()
	if p["t"] != "in" {
		t.Errorf("expected t=in, got %q", p["t"])
	}
	if p["st"] != "0" {
		t.Errorf("expected st=0, got %q", p["st"])
	}
	if p["d"] != "2.5" {
		t.Errorf("expected d=2.5, got %q", p["d"])
	}
}

func TestAudioFadeOut_NameTarget(t *testing.T) {
	e := AudioFadeOut(time.Second)
	if e.Name() != "audio_fade_out" {
		t.Errorf("expected 'audio_fade_out', got %q", e.Name())
	}
	if e.Target() != TargetAudio {
		t.Errorf("expected TargetAudio, got %v", e.Target())
	}
}

func TestAudioFadeOut_Dur(t *testing.T) {
	e := AudioFadeOut(1500 * time.Millisecond)
	if e.Dur() != 1500*time.Millisecond {
		t.Errorf("expected 1.5s, got %v", e.Dur())
	}
}

func TestAudioFadeOut_Params(t *testing.T) {
	e := AudioFadeOut(800 * time.Millisecond)
	p := e.FilterParams()
	if p["st"] != "0" {
		t.Errorf("expected st=0 for default audio fade-out, got %q", p["st"])
	}
	if p["d"] != "0.8" {
		t.Errorf("expected d=0.8, got %q", p["d"])
	}
}

func TestAudioFadeOutAt(t *testing.T) {
	e := AudioFadeOutAt(8*time.Second, 1200*time.Millisecond)
	if e.Name() != "audio_fade_out" {
		t.Errorf("expected 'audio_fade_out', got %q", e.Name())
	}
	if e.Target() != TargetAudio {
		t.Errorf("expected TargetAudio, got %v", e.Target())
	}
	if e.FilterName() != "afade" {
		t.Errorf("expected 'afade', got %q", e.FilterName())
	}
	if e.Dur() != 1200*time.Millisecond {
		t.Errorf("expected 1.2s, got %v", e.Dur())
	}
	p := e.FilterParams()
	if p["t"] != "out" {
		t.Errorf("expected t=out, got %q", p["t"])
	}
	if p["st"] != "8" {
		t.Errorf("expected st=8, got %q", p["st"])
	}
	if p["d"] != "1.2" {
		t.Errorf("expected d=1.2, got %q", p["d"])
	}
}

func TestVolume_EdgeValues(t *testing.T) {
	tests := []struct {
		level        float64
		wantVolume   string
	}{
		{0.0, "0"},
		{2.0, "2"},
		{1.0, "1"},
		{3.5, "3.5"},
	}

	for _, tt := range tests {
		e := Volume(tt.level)
		if e.Level() != tt.level {
			t.Errorf("Volume(%f).Level() = %f, want %f", tt.level, e.Level(), tt.level)
		}
		p := e.FilterParams()
		if p["volume"] != tt.wantVolume {
			t.Errorf("Volume(%f) params[volume] = %q, want %q", tt.level, p["volume"], tt.wantVolume)
		}
	}
}

func TestVolume_DurationFactor(t *testing.T) {
	e := Volume(0.25)
	if f := e.DurationFactor(); f != 1.0 {
		t.Errorf("expected DurationFactor=1.0, got %f", f)
	}
}

func TestNormalize_Target(t *testing.T) {
	e := Normalize()
	if e.Target() != TargetAudio {
		t.Errorf("expected TargetAudio, got %v", e.Target())
	}
}

func TestNormalize_Params(t *testing.T) {
	e := Normalize()
	p := e.FilterParams()
	if p["I"] != "-16" {
		t.Errorf("expected I=-16, got %q", p["I"])
	}
}

func TestNormalizeTo_Params(t *testing.T) {
	e := NormalizeTo(-14.0)
	p := e.FilterParams()
	if p["I"] != "-14" {
		t.Errorf("expected I=-14, got %q", p["I"])
	}
}

func TestNormalizeTo_Full(t *testing.T) {
	e := NormalizeTo(-20.5)
	if e.Name() != "normalize" {
		t.Errorf("expected 'normalize', got %q", e.Name())
	}
	if e.FilterName() != "loudnorm" {
		t.Errorf("expected 'loudnorm', got %q", e.FilterName())
	}
	if e.Target() != TargetAudio {
		t.Errorf("expected TargetAudio, got %v", e.Target())
	}
	if e.TargetLUFS() != -20.5 {
		t.Errorf("expected -20.5, got %f", e.TargetLUFS())
	}
}

func TestAudioSpeed_Target(t *testing.T) {
	e := AudioSpeed(2.0)
	if e.Target() != TargetAudio {
		t.Errorf("expected TargetAudio, got %v", e.Target())
	}
}

func TestAudioSpeed_Params(t *testing.T) {
	e := AudioSpeed(0.75)
	p := e.FilterParams()
	if p["tempo"] != "0.75" {
		t.Errorf("expected tempo=0.75, got %q", p["tempo"])
	}
}

func TestAudioSpeed_Zero(t *testing.T) {
	e := AudioSpeed(0)
	if e.Factor() != 0 {
		t.Errorf("expected factor 0, got %f", e.Factor())
	}
	if f := e.DurationFactor(); f != 1.0 {
		t.Errorf("expected DurationFactor=1.0 for AudioSpeed(0), got %f", f)
	}
}

func TestAudioSpeed_One(t *testing.T) {
	e := AudioSpeed(1.0)
	if e.Factor() != 1.0 {
		t.Errorf("expected factor 1.0, got %f", e.Factor())
	}
	if f := e.DurationFactor(); f != 1.0 {
		t.Errorf("expected DurationFactor=1.0, got %f", f)
	}
}

func TestFormatSeconds(t *testing.T) {
	tests := []struct {
		input time.Duration
		want  string
	}{
		{time.Second, "1"},
		{500 * time.Millisecond, "0.5"},
		{2 * time.Second, "2"},
		{1500 * time.Millisecond, "1.5"},
		{0, "0"},
	}

	for _, tt := range tests {
		got := formatSeconds(tt.input)
		if got != tt.want {
			t.Errorf("formatSeconds(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatFloat_Edge(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0.0, "0"},
		{0.1, "0.1"},
		{0.000333, "0.000333"},
		{100.0, "100"},
		{-1.0, "-1"},
	}

	for _, tt := range tests {
		got := formatFloat(tt.input)
		if got != tt.want {
			t.Errorf("formatFloat(%f) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestTargetConstants(t *testing.T) {
	if TargetVideo != 0 {
		t.Errorf("expected TargetVideo=0, got %d", TargetVideo)
	}
	if TargetAudio != 1 {
		t.Errorf("expected TargetAudio=1, got %d", TargetAudio)
	}
	if TargetBoth != 2 {
		t.Errorf("expected TargetBoth=2, got %d", TargetBoth)
	}
}
