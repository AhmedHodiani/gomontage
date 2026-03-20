package engine

import (
	"strings"
	"testing"
	"time"
)

func TestParseFFmpegTime(t *testing.T) {
	tests := []struct {
		input   string
		want    time.Duration
		wantErr bool
	}{
		{"00:00:00.00", 0, false},
		{"00:00:10.00", 10 * time.Second, false},
		{"00:01:30.50", time.Minute + 30*time.Second + 500*time.Millisecond, false},
		{"01:00:00.00", time.Hour, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		got, err := parseFFmpegTime(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseFFmpegTime(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("parseFFmpegTime(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParseProgress(t *testing.T) {
	// Simulate FFmpeg stderr output.
	input := strings.NewReader(
		"frame=  150 fps= 30.0 q=28.0 size=    2048kB time=00:00:05.00 bitrate=3355.4kbits/s speed=1.50x\r" +
			"frame=  300 fps= 29.5 q=28.0 size=    4096kB time=00:00:10.00 bitrate=3355.4kbits/s speed=1.48x\r",
	)

	var updates []Progress
	parseProgress(input, func(p Progress) {
		updates = append(updates, p)
	})

	if len(updates) != 2 {
		t.Fatalf("expected 2 progress updates, got %d", len(updates))
	}

	if updates[0].Frame != 150 {
		t.Errorf("expected frame 150, got %d", updates[0].Frame)
	}
	if updates[0].FPS != 30.0 {
		t.Errorf("expected fps 30.0, got %f", updates[0].FPS)
	}
	if updates[0].Time != 5*time.Second {
		t.Errorf("expected time 5s, got %v", updates[0].Time)
	}
	if updates[0].Speed != 1.50 {
		t.Errorf("expected speed 1.50, got %f", updates[0].Speed)
	}

	if updates[1].Frame != 300 {
		t.Errorf("expected frame 300, got %d", updates[1].Frame)
	}
}

func TestParseProgress_NilCallback(t *testing.T) {
	// Should not panic with nil callback.
	input := strings.NewReader("some ffmpeg output\n")
	parseProgress(input, nil)
}
