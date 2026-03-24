package timeline

import (
	"strings"
	"testing"
	"time"

	"github.com/ahmedhodiani/gomontage/clip"
	"github.com/ahmedhodiani/gomontage/export"
)

// --- SubRange unit tests (structure-level) ---

func TestSubRange_EmptyRange(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("input.mp4", 30*time.Second), At(0))

	// start >= end → empty timeline.
	sub := tl.SubRange(20*time.Second, 10*time.Second)
	if len(sub.VideoTracks()) != 0 {
		t.Error("expected no video tracks for inverted range")
	}

	// start == end → empty.
	sub = tl.SubRange(10*time.Second, 10*time.Second)
	if len(sub.VideoTracks()) != 0 {
		t.Error("expected no video tracks for zero-width range")
	}
}

func TestSubRange_FullRange(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	v := clip.NewVideoWithDuration("input.mp4", 30*time.Second)
	track.Add(v, At(0))

	sub := tl.SubRange(0, 30*time.Second)
	if len(sub.VideoTracks()) != 1 {
		t.Fatalf("expected 1 video track, got %d", len(sub.VideoTracks()))
	}
	entries := sub.VideoTracks()[0].Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	// Full range returns the original clip (no trimming needed).
	if entries[0].Clip != v {
		t.Error("full range should return the same clip")
	}
	if entries[0].StartAt != 0 {
		t.Errorf("expected StartAt 0, got %v", entries[0].StartAt)
	}
}

func TestSubRange_ExcludesClipBeforeRange(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("early.mp4", 10*time.Second), At(0))
	track.Add(clip.NewVideoWithDuration("late.mp4", 10*time.Second), At(20*time.Second))

	// Range starts at 15s, so the first clip (0-10s) is entirely before the range.
	sub := tl.SubRange(15*time.Second, 30*time.Second)
	entries := sub.VideoTracks()[0].Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry (early clip excluded), got %d", len(entries))
	}
	if entries[0].Clip.SourcePath() != "late.mp4" {
		t.Errorf("expected late.mp4, got %s", entries[0].Clip.SourcePath())
	}
	// late.mp4 was at 20s, range starts at 15s → shifted to 5s.
	if entries[0].StartAt != 5*time.Second {
		t.Errorf("expected StartAt 5s, got %v", entries[0].StartAt)
	}
}

func TestSubRange_ExcludesClipAfterRange(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("early.mp4", 10*time.Second), At(0))
	track.Add(clip.NewVideoWithDuration("late.mp4", 10*time.Second), At(20*time.Second))

	// Range ends at 15s, so the second clip (20-30s) is entirely after.
	sub := tl.SubRange(0, 15*time.Second)
	entries := sub.VideoTracks()[0].Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry (late clip excluded), got %d", len(entries))
	}
	if entries[0].Clip.SourcePath() != "early.mp4" {
		t.Errorf("expected early.mp4, got %s", entries[0].Clip.SourcePath())
	}
}

func TestSubRange_TrimsClipOverlappingStart(t *testing.T) {
	// Clip spans 0-30s, range is 10-30s → clip should be trimmed to show its 10-30s portion.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("input.mp4", 30*time.Second), At(0))

	sub := tl.SubRange(10*time.Second, 30*time.Second)
	entries := sub.VideoTracks()[0].Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	c := entries[0].Clip
	if c.Duration() != 20*time.Second {
		t.Errorf("expected duration 20s, got %v", c.Duration())
	}
	// The clip starts before the range → shifted to 0.
	if entries[0].StartAt != 0 {
		t.Errorf("expected StartAt 0, got %v", entries[0].StartAt)
	}
	// Source trim should reflect the sub-range: 10-30s in source coordinates.
	if c.TrimStart() != 10*time.Second {
		t.Errorf("expected TrimStart 10s, got %v", c.TrimStart())
	}
	if c.TrimEnd() != 30*time.Second {
		t.Errorf("expected TrimEnd 30s, got %v", c.TrimEnd())
	}
}

func TestSubRange_TrimsClipOverlappingEnd(t *testing.T) {
	// Clip spans 0-30s, range is 0-20s → clip should show only its 0-20s portion.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("input.mp4", 30*time.Second), At(0))

	sub := tl.SubRange(0, 20*time.Second)
	entries := sub.VideoTracks()[0].Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	c := entries[0].Clip
	if c.Duration() != 20*time.Second {
		t.Errorf("expected duration 20s, got %v", c.Duration())
	}
	if c.TrimStart() != 0 {
		t.Errorf("expected TrimStart 0, got %v", c.TrimStart())
	}
	if c.TrimEnd() != 20*time.Second {
		t.Errorf("expected TrimEnd 20s, got %v", c.TrimEnd())
	}
}

func TestSubRange_TrimsClipOnBothSides(t *testing.T) {
	// Clip spans 0-30s, range is 5-25s → clip shows only 5-25s.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("input.mp4", 30*time.Second), At(0))

	sub := tl.SubRange(5*time.Second, 25*time.Second)
	entries := sub.VideoTracks()[0].Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	c := entries[0].Clip
	if c.Duration() != 20*time.Second {
		t.Errorf("expected duration 20s, got %v", c.Duration())
	}
	if c.TrimStart() != 5*time.Second {
		t.Errorf("expected TrimStart 5s, got %v", c.TrimStart())
	}
	if c.TrimEnd() != 25*time.Second {
		t.Errorf("expected TrimEnd 25s, got %v", c.TrimEnd())
	}
}

func TestSubRange_ShiftsStartAt(t *testing.T) {
	// Clip at 20s in the timeline, range starts at 10s → clip should appear at 10s in sub.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("input.mp4", 10*time.Second), At(20*time.Second))

	sub := tl.SubRange(10*time.Second, 40*time.Second)
	entries := sub.VideoTracks()[0].Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].StartAt != 10*time.Second {
		t.Errorf("expected StartAt 10s (shifted), got %v", entries[0].StartAt)
	}
}

func TestSubRange_PreservesConfig(t *testing.T) {
	cfg := Config{Width: 3840, Height: 2160, FPS: 60}
	tl := New(cfg)
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("input.mp4", 10*time.Second), At(0))

	sub := tl.SubRange(0, 5*time.Second)
	subCfg := sub.Config()
	if subCfg.Width != cfg.Width || subCfg.Height != cfg.Height || subCfg.FPS != cfg.FPS {
		t.Errorf("expected config %+v, got %+v", cfg, subCfg)
	}
}

func TestSubRange_PreservesTrackNames(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	tl.AddVideoTrack("footage")
	tl.AddVideoTrack("overlay")
	tl.AddAudioTrack("music")

	// Add clips so tracks are not empty (or at least the track structure is preserved).
	tl.VideoTracks()[0].Add(clip.NewVideoWithDuration("a.mp4", 10*time.Second), At(0))
	tl.VideoTracks()[1].Add(clip.NewVideoWithDuration("b.mp4", 10*time.Second), At(0))
	tl.AudioTracks()[0].Add(clip.NewAudioWithDuration("m.mp3", 10*time.Second), At(0))

	sub := tl.SubRange(0, 5*time.Second)
	if len(sub.VideoTracks()) != 2 {
		t.Fatalf("expected 2 video tracks, got %d", len(sub.VideoTracks()))
	}
	if sub.VideoTracks()[0].Name() != "footage" {
		t.Errorf("expected track name 'footage', got %q", sub.VideoTracks()[0].Name())
	}
	if sub.VideoTracks()[1].Name() != "overlay" {
		t.Errorf("expected track name 'overlay', got %q", sub.VideoTracks()[1].Name())
	}
	if len(sub.AudioTracks()) != 1 {
		t.Fatalf("expected 1 audio track, got %d", len(sub.AudioTracks()))
	}
	if sub.AudioTracks()[0].Name() != "music" {
		t.Errorf("expected track name 'music', got %q", sub.AudioTracks()[0].Name())
	}
}

func TestSubRange_AudioTrackFiltered(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	vTrack := tl.AddVideoTrack("main")
	aTrack := tl.AddAudioTrack("music")

	vTrack.Add(clip.NewVideoWithDuration("footage.mp4", 60*time.Second), At(0))
	aTrack.Add(clip.NewAudioWithDuration("short.mp3", 10*time.Second), At(0))
	aTrack.Add(clip.NewAudioWithDuration("long.mp3", 30*time.Second), At(20*time.Second))

	// Range 15-50s: short.mp3 (0-10s) is entirely before → excluded.
	// long.mp3 (20-50s) overlaps → trimmed and shifted.
	sub := tl.SubRange(15*time.Second, 50*time.Second)
	audioEntries := sub.AudioTracks()[0].Entries()
	if len(audioEntries) != 1 {
		t.Fatalf("expected 1 audio entry, got %d", len(audioEntries))
	}
	if audioEntries[0].Clip.SourcePath() != "long.mp3" {
		t.Errorf("expected long.mp3, got %s", audioEntries[0].Clip.SourcePath())
	}
	// long.mp3 was at 20s, range starts at 15s → shifted to 5s.
	if audioEntries[0].StartAt != 5*time.Second {
		t.Errorf("expected StartAt 5s, got %v", audioEntries[0].StartAt)
	}
	if audioEntries[0].Clip.Duration() != 30*time.Second {
		t.Errorf("expected full 30s duration (clip within range), got %v", audioEntries[0].Clip.Duration())
	}
}

func TestSubRange_NegativeStartClamped(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("input.mp4", 30*time.Second), At(0))

	sub := tl.SubRange(-10*time.Second, 20*time.Second)
	entries := sub.VideoTracks()[0].Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	// Negative start is clamped to 0, so the clip is trimmed to 0-20s.
	if entries[0].Clip.Duration() != 20*time.Second {
		t.Errorf("expected duration 20s, got %v", entries[0].Clip.Duration())
	}
}

func TestSubRange_EndPastDuration(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("input.mp4", 30*time.Second), At(0))

	sub := tl.SubRange(10*time.Second, 999*time.Second)
	entries := sub.VideoTracks()[0].Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	// End clamped to 30s, so clip shows 10-30s = 20s.
	if entries[0].Clip.Duration() != 20*time.Second {
		t.Errorf("expected duration 20s, got %v", entries[0].Clip.Duration())
	}
}

func TestSubRange_MultipleClipsPartialOverlap(t *testing.T) {
	// Three sequential clips: 0-10s, 10-20s, 20-30s.
	// Range: 5-25s → all three partially overlap.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("a.mp4", 10*time.Second), At(0))
	track.Add(clip.NewVideoWithDuration("b.mp4", 10*time.Second), At(10*time.Second))
	track.Add(clip.NewVideoWithDuration("c.mp4", 10*time.Second), At(20*time.Second))

	sub := tl.SubRange(5*time.Second, 25*time.Second)
	entries := sub.VideoTracks()[0].Entries()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	// a.mp4: originally 0-10s, visible 5-10s → 5s duration, starts at 0.
	if entries[0].Clip.Duration() != 5*time.Second {
		t.Errorf("clip a: expected duration 5s, got %v", entries[0].Clip.Duration())
	}
	if entries[0].StartAt != 0 {
		t.Errorf("clip a: expected StartAt 0, got %v", entries[0].StartAt)
	}

	// b.mp4: originally 10-20s, fully within range → 10s duration, starts at 5s.
	if entries[1].Clip.Duration() != 10*time.Second {
		t.Errorf("clip b: expected duration 10s, got %v", entries[1].Clip.Duration())
	}
	if entries[1].StartAt != 5*time.Second {
		t.Errorf("clip b: expected StartAt 5s, got %v", entries[1].StartAt)
	}

	// c.mp4: originally 20-30s, visible 20-25s → 5s duration, starts at 15s.
	if entries[2].Clip.Duration() != 5*time.Second {
		t.Errorf("clip c: expected duration 5s, got %v", entries[2].Clip.Duration())
	}
	if entries[2].StartAt != 15*time.Second {
		t.Errorf("clip c: expected StartAt 15s, got %v", entries[2].StartAt)
	}
}

func TestSubRange_GapPreserved(t *testing.T) {
	// Clip at 0-10s, gap 10-20s, clip at 20-30s.
	// Range: 0-30s (full) → gap is preserved.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("a.mp4", 10*time.Second), At(0))
	track.Add(clip.NewVideoWithDuration("b.mp4", 10*time.Second), At(20*time.Second))

	sub := tl.SubRange(5*time.Second, 25*time.Second)
	entries := sub.VideoTracks()[0].Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// a.mp4: 0-10s, visible 5-10s → 5s, starts at 0.
	if entries[0].StartAt != 0 {
		t.Errorf("clip a: expected StartAt 0, got %v", entries[0].StartAt)
	}
	if entries[0].Clip.Duration() != 5*time.Second {
		t.Errorf("clip a: expected 5s, got %v", entries[0].Clip.Duration())
	}

	// b.mp4: 20-30s, visible 20-25s → 5s, starts at 15s (20-5).
	if entries[1].StartAt != 15*time.Second {
		t.Errorf("clip b: expected StartAt 15s (gap preserved), got %v", entries[1].StartAt)
	}
	if entries[1].Clip.Duration() != 5*time.Second {
		t.Errorf("clip b: expected 5s, got %v", entries[1].Clip.Duration())
	}
}

func TestSubRange_ImageClipTrimmed(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	img := clip.NewImage("bg.png").WithDuration(20*time.Second).WithSize(1920, 1080)
	track.Add(img, At(0))

	sub := tl.SubRange(5*time.Second, 15*time.Second)
	entries := sub.VideoTracks()[0].Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Clip.Duration() != 10*time.Second {
		t.Errorf("expected image clip trimmed to 10s, got %v", entries[0].Clip.Duration())
	}
}

func TestSubRange_ColorClipTrimmed(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	c := clip.NewColor("#000000", 1920, 1080).WithDuration(10 * time.Second)
	track.Add(c, At(0))

	sub := tl.SubRange(2*time.Second, 8*time.Second)
	entries := sub.VideoTracks()[0].Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Clip.Duration() != 6*time.Second {
		t.Errorf("expected color clip trimmed to 6s, got %v", entries[0].Clip.Duration())
	}
}

// --- SubRange + DryRun integration tests ---

func TestDryRunRange_BasicTrim(t *testing.T) {
	// DryRunRange should produce a trimmed command.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("input.mp4", 60*time.Second), At(0))

	cmd, err := tl.DryRunRange(10*time.Second, 20*time.Second, export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRunRange returned error: %v", err)
	}

	cmdStr := cmd.String()
	if !strings.Contains(cmdStr, "input.mp4") {
		t.Error("expected command to contain input file")
	}
	if !strings.Contains(cmdStr, "output.mp4") {
		t.Error("expected command to contain output file")
	}
	// SubRange trims the clip → should have a trim filter.
	if !strings.Contains(cmdStr, "trim") {
		t.Errorf("expected trim filter for sub-ranged clip, got:\n%s", cmdStr)
	}
}

func TestDryRunRange_ExcludesClip(t *testing.T) {
	// Range that excludes the first clip should not reference its source file
	// in the FFmpeg command.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("excluded.mp4", 10*time.Second), At(0))
	track.Add(clip.NewVideoWithDuration("included.mp4", 10*time.Second), At(20*time.Second))

	cmd, err := tl.DryRunRange(15*time.Second, 30*time.Second, export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRunRange returned error: %v", err)
	}

	cmdStr := cmd.String()
	if strings.Contains(cmdStr, "excluded.mp4") {
		t.Errorf("excluded clip should not appear in command, got:\n%s", cmdStr)
	}
	if !strings.Contains(cmdStr, "included.mp4") {
		t.Errorf("expected included clip in command, got:\n%s", cmdStr)
	}
}

func TestDryRunRange_PreAlreadyTrimmedClip(t *testing.T) {
	// Clip is already trimmed (source 20-40s) and placed at 0s.
	// SubRange(5s, 15s) should produce source coordinates 25-35s.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	v := clip.NewVideoWithDuration("source.mp4", 60*time.Second)
	trimmed := v.Trim(20*time.Second, 40*time.Second) // 20s local duration, source 20-40
	track.Add(trimmed, At(0))

	cmd, err := tl.DryRunRange(5*time.Second, 15*time.Second, export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRunRange returned error: %v", err)
	}

	cmdStr := cmd.String()
	// Trim filter should have start=25 (source coordinate).
	if !strings.Contains(cmdStr, "start=25.000") {
		t.Errorf("expected trim start=25.000 (source coordinate), got:\n%s", cmdStr)
	}
	// Trim filter should have end=35 (source coordinate).
	if !strings.Contains(cmdStr, "end=35.000") {
		t.Errorf("expected trim end=35.000 (source coordinate), got:\n%s", cmdStr)
	}
}

func TestDryRunRange_SequentialClipsMiddle(t *testing.T) {
	// Three 10s clips in sequence (0-10, 10-20, 20-30).
	// SubRange(5, 25) keeps all three but trims first and last.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.AddSequence(
		clip.NewVideoWithDuration("a.mp4", 10*time.Second).VideoOnly(),
		clip.NewVideoWithDuration("b.mp4", 10*time.Second).VideoOnly(),
		clip.NewVideoWithDuration("c.mp4", 10*time.Second).VideoOnly(),
	)

	cmd, err := tl.DryRunRange(5*time.Second, 25*time.Second, export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRunRange returned error: %v", err)
	}

	cmdStr := cmd.String()
	// All three inputs should be present.
	if !strings.Contains(cmdStr, "a.mp4") {
		t.Errorf("expected a.mp4 in command, got:\n%s", cmdStr)
	}
	if !strings.Contains(cmdStr, "b.mp4") {
		t.Errorf("expected b.mp4 in command, got:\n%s", cmdStr)
	}
	if !strings.Contains(cmdStr, "c.mp4") {
		t.Errorf("expected c.mp4 in command, got:\n%s", cmdStr)
	}
	// Should have concat since there are 3 clips.
	if !strings.Contains(cmdStr, "concat") {
		t.Errorf("expected concat filter, got:\n%s", cmdStr)
	}
}

func TestDryRunRange_WithAudioTrack(t *testing.T) {
	// SubRange should filter both video and audio tracks.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	vTrack := tl.AddVideoTrack("main")
	aTrack := tl.AddAudioTrack("music")

	vTrack.Add(clip.NewVideoWithDuration("footage.mp4", 60*time.Second), At(0))
	aTrack.Add(clip.NewAudioWithDuration("bgm.mp3", 60*time.Second), At(0))

	cmd, err := tl.DryRunRange(10*time.Second, 30*time.Second, export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRunRange returned error: %v", err)
	}

	cmdStr := cmd.String()
	if !strings.Contains(cmdStr, "footage.mp4") {
		t.Errorf("expected video input, got:\n%s", cmdStr)
	}
	if !strings.Contains(cmdStr, "bgm.mp3") {
		t.Errorf("expected audio input, got:\n%s", cmdStr)
	}
	// Both should be trimmed.
	if !strings.Contains(cmdStr, "trim") {
		t.Errorf("expected trim filter for video, got:\n%s", cmdStr)
	}
	if !strings.Contains(cmdStr, "atrim") {
		t.Errorf("expected atrim filter for audio, got:\n%s", cmdStr)
	}
}

func TestSubRange_Duration(t *testing.T) {
	// The sub-ranged timeline's Duration() should reflect the range.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("input.mp4", 60*time.Second), At(0))

	sub := tl.SubRange(10*time.Second, 30*time.Second)
	if sub.Duration() != 20*time.Second {
		t.Errorf("expected sub timeline duration 20s, got %v", sub.Duration())
	}
}

func TestSubRange_DoesNotMutateOriginal(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("input.mp4", 60*time.Second), At(0))

	origDur := tl.Duration()
	origEntries := len(tl.VideoTracks()[0].Entries())

	_ = tl.SubRange(10*time.Second, 20*time.Second)

	if tl.Duration() != origDur {
		t.Errorf("SubRange mutated original timeline duration: %v → %v", origDur, tl.Duration())
	}
	if len(tl.VideoTracks()[0].Entries()) != origEntries {
		t.Error("SubRange mutated original timeline entries")
	}
}

func TestSubRange_RangeEntirelyInGap(t *testing.T) {
	// Clip at 0-10s, gap 10-20s, clip at 20-30s.
	// Range: 12-18s (entirely in gap) → empty sub-timeline.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("a.mp4", 10*time.Second), At(0))
	track.Add(clip.NewVideoWithDuration("b.mp4", 10*time.Second), At(20*time.Second))

	sub := tl.SubRange(12*time.Second, 18*time.Second)
	// Track should exist but have no entries.
	if len(sub.VideoTracks()) != 1 {
		t.Fatalf("expected 1 video track, got %d", len(sub.VideoTracks()))
	}
	entries := sub.VideoTracks()[0].Entries()
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for range entirely in gap, got %d", len(entries))
	}
}
