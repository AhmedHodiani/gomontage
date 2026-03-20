package timeline

import (
	"strings"
	"testing"
	"time"

	"github.com/ahmedhodiani/gomontage/clip"
	"github.com/ahmedhodiani/gomontage/export"
)

func TestDryRun_BasicVideoTrack(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("input.mp4", 10*time.Second), At(0))

	cmd, err := tl.DryRun(export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRun returned error: %v", err)
	}

	cmdStr := cmd.String()
	if !strings.Contains(cmdStr, "ffmpeg") {
		t.Error("expected command to start with ffmpeg")
	}
	if !strings.Contains(cmdStr, "-i") {
		t.Error("expected command to contain -i flag")
	}
	if !strings.Contains(cmdStr, "input.mp4") {
		t.Error("expected command to contain input file")
	}
	if !strings.Contains(cmdStr, "output.mp4") {
		t.Error("expected command to contain output file")
	}
}

func TestDryRun_ProfileParams(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("input.mp4", 10*time.Second), At(0))

	cmd, err := tl.DryRun(export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRun returned error: %v", err)
	}

	cmdStr := cmd.String()
	// YouTube1080p uses libx264
	if !strings.Contains(cmdStr, "libx264") {
		t.Error("expected YouTube1080p profile to include libx264 codec")
	}
}

func TestDryRun_InvalidTimeline(t *testing.T) {
	// Empty timeline should fail validation.
	tl := New(Config{Width: 0, Height: 0, FPS: 30})
	_, err := tl.DryRun(export.YouTube1080p(), "output.mp4")
	if err == nil {
		t.Error("expected error for invalid timeline")
	}
	if !strings.Contains(err.Error(), "compilation failed") {
		t.Errorf("expected compilation failed error, got: %v", err)
	}
}

func TestDryRun_MultipleInputs(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("clip1.mp4", 10*time.Second), At(0))
	track.Add(clip.NewVideoWithDuration("clip2.mp4", 10*time.Second), At(10*time.Second))

	cmd, err := tl.DryRun(export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRun returned error: %v", err)
	}

	cmdStr := cmd.String()
	if !strings.Contains(cmdStr, "clip1.mp4") {
		t.Error("expected command to contain first input")
	}
	if !strings.Contains(cmdStr, "clip2.mp4") {
		t.Error("expected command to contain second input")
	}
}

func TestDryRun_AudioProfile(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("input.mp4", 10*time.Second), At(0))

	cmd, err := tl.DryRun(export.MP3(), "output.mp3")
	if err != nil {
		t.Fatalf("DryRun returned error: %v", err)
	}

	cmdStr := cmd.String()
	if !strings.Contains(cmdStr, "output.mp3") {
		t.Error("expected command to contain mp3 output path")
	}
}

func TestDryRun_WithAudioTrack(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	video := tl.AddVideoTrack("main")
	audio := tl.AddAudioTrack("music")

	video.Add(clip.NewVideoWithDuration("footage.mp4", 30*time.Second), At(0))
	audio.Add(clip.NewAudioWithDuration("bgm.mp3", 30*time.Second), At(0))

	cmd, err := tl.DryRun(export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRun returned error: %v", err)
	}

	cmdStr := cmd.String()
	if !strings.Contains(cmdStr, "footage.mp4") {
		t.Error("expected command to contain video input")
	}
	if !strings.Contains(cmdStr, "bgm.mp3") {
		t.Error("expected command to contain audio input")
	}
}

func TestExport_InvalidTimeline(t *testing.T) {
	// Export on an invalid timeline should return a compilation error.
	tl := New(Config{Width: 0, Height: 0, FPS: 30})
	err := tl.Export(export.YouTube1080p(), "output.mp4")
	if err == nil {
		t.Error("expected error for export on invalid timeline")
	}
}
