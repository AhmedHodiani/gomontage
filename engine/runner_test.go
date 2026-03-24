package engine

import (
	"fmt"
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
	parseProgressOpts(input, RunOptions{
		OnProgress: func(p Progress) {
			updates = append(updates, p)
		},
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
	parseProgressOpts(input, RunOptions{})
}

func TestParseProgress_StderrCapture(t *testing.T) {
	// Non-progress lines should be captured in the returned stderr tail.
	input := strings.NewReader(
		"Stream mapping:\n" +
			"  Stream #0:0 -> #0:0 (h264 -> libx264)\n" +
			"frame=  100 fps= 30.0 q=28.0 size=    1024kB time=00:00:03.33 bitrate=2515.4kbits/s speed=1.00x\r" +
			"[libx264 @ 0x55] height not divisible by 2 (1920x1081)\n" +
			"Error while opening encoder for output stream #0:0\n",
	)

	tail := parseProgressOpts(input, RunOptions{})
	if len(tail) != 4 {
		t.Fatalf("expected 4 stderr lines, got %d: %v", len(tail), tail)
	}

	// The error message should be captured.
	found := false
	for _, line := range tail {
		if strings.Contains(line, "height not divisible by 2") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected FFmpeg error in stderr tail, got: %v", tail)
	}
}

func TestParseProgress_StderrCaptureWithCallback(t *testing.T) {
	// Even when a progress callback is set, non-progress lines should be captured.
	input := strings.NewReader(
		"frame=  100 fps= 30.0 q=28.0 size=    1024kB time=00:00:03.33 bitrate=2515.4kbits/s speed=1.00x\r" +
			"[error] something went wrong\n",
	)

	var updates []Progress
	tail := parseProgressOpts(input, RunOptions{
		OnProgress: func(p Progress) {
			updates = append(updates, p)
		},
	})

	if len(updates) != 1 {
		t.Fatalf("expected 1 progress update, got %d", len(updates))
	}
	if len(tail) != 1 {
		t.Fatalf("expected 1 stderr line, got %d: %v", len(tail), tail)
	}
	if !strings.Contains(tail[0], "something went wrong") {
		t.Errorf("expected error line in stderr tail, got: %v", tail)
	}
}

func TestParseProgress_PercentAndETA(t *testing.T) {
	// When TotalDuration is set, Percent and ETA should be computed.
	input := strings.NewReader(
		"frame=  150 fps= 30.0 q=28.0 size=    2048kB time=00:00:05.00 bitrate=3355.4kbits/s speed=2.00x\r",
	)

	var updates []Progress
	parseProgressOpts(input, RunOptions{
		OnProgress: func(p Progress) {
			updates = append(updates, p)
		},
		TotalDuration: 10 * time.Second,
	})

	if len(updates) != 1 {
		t.Fatalf("expected 1 progress update, got %d", len(updates))
	}

	p := updates[0]
	if p.Percent != 50.0 {
		t.Errorf("expected Percent=50.0, got %f", p.Percent)
	}
	// ETA: 5s remaining at 2x speed = 2.5s
	expectedETA := 2500 * time.Millisecond
	if p.ETA < expectedETA-100*time.Millisecond || p.ETA > expectedETA+100*time.Millisecond {
		t.Errorf("expected ETA ~2.5s, got %v", p.ETA)
	}
}

func TestFormatDurationShort(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "00:00"},
		{30 * time.Second, "00:30"},
		{90 * time.Second, "01:30"},
		{time.Hour + 5*time.Minute + 30*time.Second, "01:05:30"},
		{-time.Second, "00:00"},
	}
	for _, tt := range tests {
		got := FormatDurationShort(tt.d)
		if got != tt.want {
			t.Errorf("FormatDurationShort(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestFormatDurationHuman(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "0s"},
		{30 * time.Second, "30s"},
		{90 * time.Second, "1m30s"},
		{5 * time.Minute, "5m"},
		{time.Hour + 5*time.Minute, "1h5m"},
		{2 * time.Hour, "2h"},
	}
	for _, tt := range tests {
		got := formatDurationHuman(tt.d)
		if got != tt.want {
			t.Errorf("formatDurationHuman(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{134217728, "128.0 MB"},
		{1073741824, "1.0 GB"},
		{-1, "0 B"},
	}
	for _, tt := range tests {
		got := formatFileSize(tt.bytes)
		if got != tt.want {
			t.Errorf("formatFileSize(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}

func TestRenderBar(t *testing.T) {
	bar := renderBar(50, 10)
	if bar != "█████░░░░░" {
		t.Errorf("renderBar(50, 10) = %q", bar)
	}

	bar = renderBar(0, 10)
	if bar != "░░░░░░░░░░" {
		t.Errorf("renderBar(0, 10) = %q", bar)
	}

	bar = renderBar(100, 10)
	if bar != "██████████" {
		t.Errorf("renderBar(100, 10) = %q", bar)
	}
}

func TestFFmpegError(t *testing.T) {
	err := newFFmpegError(
		fmt.Errorf("exit status 1"),
		[]string{"[libx264] height not divisible by 2", "Error while opening encoder"},
	)

	errStr := err.Error()
	if !strings.Contains(errStr, "ffmpeg failed") {
		t.Error("expected 'ffmpeg failed' in error message")
	}
	if !strings.Contains(errStr, "height not divisible by 2") {
		t.Error("expected stderr content in error message")
	}
	if !strings.Contains(errStr, "Error while opening encoder") {
		t.Error("expected stderr content in error message")
	}

	// Unwrap should return the original error.
	if err.Unwrap() == nil {
		t.Error("expected Unwrap to return the original error")
	}
}

func TestFFmpegError_NoStderr(t *testing.T) {
	err := newFFmpegError(fmt.Errorf("exit status 1"), nil)
	errStr := err.Error()
	if !strings.Contains(errStr, "ffmpeg failed: exit status 1") {
		t.Errorf("expected simple error message, got: %q", errStr)
	}
}

func TestAppendRing(t *testing.T) {
	var buf []string

	// Fill the buffer.
	for i := 0; i < 5; i++ {
		buf = appendRing(buf, fmt.Sprintf("line%d", i), 3)
	}

	// Should only keep the last 3.
	if len(buf) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(buf))
	}
	if buf[0] != "line2" || buf[1] != "line3" || buf[2] != "line4" {
		t.Errorf("expected [line2, line3, line4], got %v", buf)
	}
}
