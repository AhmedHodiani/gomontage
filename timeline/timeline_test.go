package timeline

import (
	"testing"
	"time"

	"github.com/ahmedhodiani/gomontage/clip"
)

func TestNew(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})

	if tl.Config().Width != 1920 {
		t.Errorf("expected width 1920, got %d", tl.Config().Width)
	}
	if tl.Config().Height != 1080 {
		t.Errorf("expected height 1080, got %d", tl.Config().Height)
	}
	if tl.Config().FPS != 30 {
		t.Errorf("expected FPS 30, got %f", tl.Config().FPS)
	}
}

func TestNew_DefaultFPS(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080})
	if tl.Config().FPS != 30 {
		t.Errorf("expected default FPS 30, got %f", tl.Config().FPS)
	}
}

func TestTimeline_AddTracks(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})

	v1 := tl.AddVideoTrack("main")
	v2 := tl.AddVideoTrack("overlay")
	a1 := tl.AddAudioTrack("narration")
	a2 := tl.AddAudioTrack("music")

	if len(tl.VideoTracks()) != 2 {
		t.Errorf("expected 2 video tracks, got %d", len(tl.VideoTracks()))
	}
	if len(tl.AudioTracks()) != 2 {
		t.Errorf("expected 2 audio tracks, got %d", len(tl.AudioTracks()))
	}
	if v1.Name() != "main" {
		t.Errorf("expected name 'main', got %q", v1.Name())
	}
	if v2.Index() != 1 {
		t.Errorf("expected index 1, got %d", v2.Index())
	}
	if a1.Name() != "narration" {
		t.Errorf("expected name 'narration', got %q", a1.Name())
	}
	if a2.Index() != 1 {
		t.Errorf("expected index 1, got %d", a2.Index())
	}
}

func TestVideoTrack_Add(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")

	c1 := clip.NewVideoWithDuration("a.mp4", 10*time.Second)
	c2 := clip.NewVideoWithDuration("b.mp4", 15*time.Second)

	track.Add(c1, At(0))
	track.Add(c2, At(10*time.Second))

	entries := track.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].StartAt != 0 {
		t.Errorf("expected start 0, got %v", entries[0].StartAt)
	}
	if entries[1].StartAt != 10*time.Second {
		t.Errorf("expected start 10s, got %v", entries[1].StartAt)
	}
}

func TestVideoTrack_AddSequence(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")

	c1 := clip.NewVideoWithDuration("a.mp4", 10*time.Second)
	c2 := clip.NewVideoWithDuration("b.mp4", 15*time.Second)
	c3 := clip.NewVideoWithDuration("c.mp4", 5*time.Second)

	track.AddSequence(c1, c2, c3)

	entries := track.Entries()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].StartAt != 0 {
		t.Errorf("clip 1 should start at 0, got %v", entries[0].StartAt)
	}
	if entries[1].StartAt != 10*time.Second {
		t.Errorf("clip 2 should start at 10s, got %v", entries[1].StartAt)
	}
	if entries[2].StartAt != 25*time.Second {
		t.Errorf("clip 3 should start at 25s, got %v", entries[2].StartAt)
	}
}

func TestVideoTrack_End(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")

	if track.End() != 0 {
		t.Errorf("empty track should have end 0, got %v", track.End())
	}

	c := clip.NewVideoWithDuration("a.mp4", 10*time.Second)
	track.Add(c, At(5*time.Second))

	if track.End() != 15*time.Second {
		t.Errorf("expected end 15s, got %v", track.End())
	}
}

func TestAudioTrack_AddSequence(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddAudioTrack("voice")

	a1 := clip.NewAudioWithDuration("ch1.wav", 30*time.Second)
	a2 := clip.NewAudioWithDuration("ch2.wav", 45*time.Second)

	track.AddSequence(a1, a2)

	entries := track.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[1].StartAt != 30*time.Second {
		t.Errorf("expected ch2 at 30s, got %v", entries[1].StartAt)
	}
	if track.End() != 75*time.Second {
		t.Errorf("expected end 75s, got %v", track.End())
	}
}

func TestTimeline_Duration(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})

	video := tl.AddVideoTrack("main")
	audio := tl.AddAudioTrack("music")

	video.Add(clip.NewVideoWithDuration("v.mp4", 30*time.Second), At(0))
	audio.Add(clip.NewAudioWithDuration("m.mp3", 60*time.Second), At(0))

	if tl.Duration() != 60*time.Second {
		t.Errorf("expected 60s (longest track), got %v", tl.Duration())
	}
}

func TestTimeline_Validate(t *testing.T) {
	// Invalid resolution.
	tl := New(Config{Width: 0, Height: 0, FPS: 30})
	if err := tl.Validate(); err == nil {
		t.Error("expected validation error for invalid resolution")
	}

	// No tracks.
	tl = New(Config{Width: 1920, Height: 1080, FPS: 30})
	if err := tl.Validate(); err == nil {
		t.Error("expected validation error for no tracks")
	}

	// Track but no clips.
	tl = New(Config{Width: 1920, Height: 1080, FPS: 30})
	tl.AddVideoTrack("empty")
	if err := tl.Validate(); err == nil {
		t.Error("expected validation error for no clips")
	}

	// Valid timeline.
	tl = New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("v.mp4", 10*time.Second), At(0))
	if err := tl.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Overlapping clips on same video track should fail.
	tl = New(Config{Width: 1920, Height: 1080, FPS: 30})
	track = tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("a.mp4", 10*time.Second), At(0))
	track.Add(clip.NewVideoWithDuration("b.mp4", 10*time.Second), At(5*time.Second)) // overlaps by 5s
	if err := tl.Validate(); err == nil {
		t.Error("expected validation error for overlapping clips on same track")
	}

	// Adjacent clips (no overlap) should pass.
	tl = New(Config{Width: 1920, Height: 1080, FPS: 30})
	track = tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("a.mp4", 10*time.Second), At(0))
	track.Add(clip.NewVideoWithDuration("b.mp4", 10*time.Second), At(10*time.Second)) // exactly adjacent
	if err := tl.Validate(); err != nil {
		t.Errorf("expected no error for adjacent clips, got %v", err)
	}
}

func TestFormatSeconds(t *testing.T) {
	tests := []struct {
		input time.Duration
		want  string
	}{
		{0, "0.000"},
		{5 * time.Second, "5.000"},
		{1500 * time.Millisecond, "1.500"},
		{time.Minute + 30*time.Second, "90.000"},
	}

	for _, tt := range tests {
		got := formatSeconds(tt.input)
		if got != tt.want {
			t.Errorf("formatSeconds(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
