package gomontage

import (
	"testing"
	"time"
)

func TestNewTimeline(t *testing.T) {
	tl := NewTimeline(TimelineConfig{
		Width:  1920,
		Height: 1080,
		FPS:    30,
	})

	cfg := tl.Config()
	if cfg.Width != 1920 || cfg.Height != 1080 || cfg.FPS != 30 {
		t.Errorf("unexpected config: %+v", cfg)
	}
}

func TestAt(t *testing.T) {
	result := At(5 * time.Second)
	if result != 5*time.Second {
		t.Errorf("At(5s) = %v, want 5s", result)
	}
}

func TestHD(t *testing.T) {
	cfg := HD()
	if cfg.Width != 1920 || cfg.Height != 1080 || cfg.FPS != 30 {
		t.Errorf("HD() = %+v, want 1920x1080@30", cfg)
	}
}

func TestHD60(t *testing.T) {
	cfg := HD60()
	if cfg.Width != 1920 || cfg.Height != 1080 || cfg.FPS != 60 {
		t.Errorf("HD60() = %+v, want 1920x1080@60", cfg)
	}
}

func TestUHD(t *testing.T) {
	cfg := UHD()
	if cfg.Width != 3840 || cfg.Height != 2160 || cfg.FPS != 30 {
		t.Errorf("UHD() = %+v, want 3840x2160@30", cfg)
	}
}

func TestUHD60(t *testing.T) {
	cfg := UHD60()
	if cfg.Width != 3840 || cfg.Height != 2160 || cfg.FPS != 60 {
		t.Errorf("UHD60() = %+v, want 3840x2160@60", cfg)
	}
}

func TestVertical(t *testing.T) {
	cfg := Vertical()
	if cfg.Width != 1080 || cfg.Height != 1920 || cfg.FPS != 30 {
		t.Errorf("Vertical() = %+v, want 1080x1920@30", cfg)
	}
}

func TestSquare(t *testing.T) {
	cfg := Square()
	if cfg.Width != 1080 || cfg.Height != 1080 || cfg.FPS != 30 {
		t.Errorf("Square() = %+v, want 1080x1080@30", cfg)
	}
}

func TestSeconds(t *testing.T) {
	tests := []struct {
		input float64
		want  time.Duration
	}{
		{1.0, 1 * time.Second},
		{0.5, 500 * time.Millisecond},
		{5.5, 5*time.Second + 500*time.Millisecond},
		{0.0, 0},
		{60.0, 1 * time.Minute},
	}

	for _, tt := range tests {
		got := Seconds(tt.input)
		if got != tt.want {
			t.Errorf("Seconds(%v) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestMinutes(t *testing.T) {
	tests := []struct {
		input float64
		want  time.Duration
	}{
		{1.0, 1 * time.Minute},
		{2.5, 2*time.Minute + 30*time.Second},
		{0.0, 0},
	}

	for _, tt := range tests {
		got := Minutes(tt.input)
		if got != tt.want {
			t.Errorf("Minutes(%v) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
