package timeline

import (
	"strings"
	"testing"
	"time"

	"github.com/ahmedhodiani/gomontage/clip"
	"github.com/ahmedhodiani/gomontage/effects"
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

// --- Timeline-level edge case tests ---

func TestSubRange_MultiTrackDifferentFiltering(t *testing.T) {
	// Two video tracks with clips at different positions.
	// Track 1: clip at 0-10s (excluded by range).
	// Track 2: clip at 20-30s (included).
	// Range: 15-30s.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track1 := tl.AddVideoTrack("background")
	track2 := tl.AddVideoTrack("overlay")

	track1.Add(clip.NewVideoWithDuration("bg.mp4", 10*time.Second), At(0))
	track2.Add(clip.NewVideoWithDuration("overlay.mp4", 10*time.Second), At(20*time.Second))

	sub := tl.SubRange(15*time.Second, 30*time.Second)

	// Track 1 should have 0 entries (clip excluded).
	if len(sub.VideoTracks()[0].Entries()) != 0 {
		t.Errorf("track1: expected 0 entries, got %d", len(sub.VideoTracks()[0].Entries()))
	}
	// Track 2 should have 1 entry (clip included and shifted).
	entries2 := sub.VideoTracks()[1].Entries()
	if len(entries2) != 1 {
		t.Fatalf("track2: expected 1 entry, got %d", len(entries2))
	}
	if entries2[0].StartAt != 5*time.Second {
		t.Errorf("track2: expected StartAt 5s (20-15), got %v", entries2[0].StartAt)
	}
	if entries2[0].Clip.Duration() != 10*time.Second {
		t.Errorf("track2: expected full 10s duration, got %v", entries2[0].Clip.Duration())
	}
}

func TestSubRange_MultiTrackPartialTrimming(t *testing.T) {
	// Track 1: clip at 0-20s (partially trimmed).
	// Track 2: clip at 10-30s (partially trimmed differently).
	// Range: 5-25s.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track1 := tl.AddVideoTrack("bg")
	track2 := tl.AddVideoTrack("fg")

	track1.Add(clip.NewVideoWithDuration("bg.mp4", 20*time.Second), At(0))
	track2.Add(clip.NewVideoWithDuration("fg.mp4", 20*time.Second), At(10*time.Second))

	sub := tl.SubRange(5*time.Second, 25*time.Second)

	// Track 1: clip 0-20s, visible 5-20s → 15s, starts at 0.
	entries1 := sub.VideoTracks()[0].Entries()
	if len(entries1) != 1 {
		t.Fatalf("track1: expected 1 entry, got %d", len(entries1))
	}
	if entries1[0].Clip.Duration() != 15*time.Second {
		t.Errorf("track1: expected duration 15s, got %v", entries1[0].Clip.Duration())
	}
	if entries1[0].StartAt != 0 {
		t.Errorf("track1: expected StartAt 0, got %v", entries1[0].StartAt)
	}

	// Track 2: clip 10-30s, visible 10-25s → 15s, starts at 5s.
	entries2 := sub.VideoTracks()[1].Entries()
	if len(entries2) != 1 {
		t.Fatalf("track2: expected 1 entry, got %d", len(entries2))
	}
	if entries2[0].Clip.Duration() != 15*time.Second {
		t.Errorf("track2: expected duration 15s, got %v", entries2[0].Clip.Duration())
	}
	if entries2[0].StartAt != 5*time.Second {
		t.Errorf("track2: expected StartAt 5s, got %v", entries2[0].StartAt)
	}
}

func TestSubRange_OverlappingAudioClips(t *testing.T) {
	// Audio tracks allow overlapping clips. Two overlapping audio clips:
	// music: 0-30s, narration: 10-40s.
	// Range: 5-35s → both trimmed independently.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	vTrack := tl.AddVideoTrack("main")
	vTrack.Add(clip.NewVideoWithDuration("video.mp4", 40*time.Second).VideoOnly(), At(0))

	aTrack := tl.AddAudioTrack("mixed")
	aTrack.Add(clip.NewAudioWithDuration("music.mp3", 30*time.Second), At(0))
	aTrack.Add(clip.NewAudioWithDuration("narration.wav", 30*time.Second), At(10*time.Second))

	sub := tl.SubRange(5*time.Second, 35*time.Second)
	audioEntries := sub.AudioTracks()[0].Entries()
	if len(audioEntries) != 2 {
		t.Fatalf("expected 2 overlapping audio entries, got %d", len(audioEntries))
	}

	// music: 0-30s, visible 5-30s → 25s, starts at 0.
	if audioEntries[0].Clip.Duration() != 25*time.Second {
		t.Errorf("music: expected 25s, got %v", audioEntries[0].Clip.Duration())
	}
	if audioEntries[0].StartAt != 0 {
		t.Errorf("music: expected StartAt 0, got %v", audioEntries[0].StartAt)
	}

	// narration: 10-40s, visible 10-35s → 25s, starts at 5s.
	if audioEntries[1].Clip.Duration() != 25*time.Second {
		t.Errorf("narration: expected 25s, got %v", audioEntries[1].Clip.Duration())
	}
	if audioEntries[1].StartAt != 5*time.Second {
		t.Errorf("narration: expected StartAt 5s, got %v", audioEntries[1].StartAt)
	}
}

func TestSubRange_NonZeroPlacementTrimAndShift(t *testing.T) {
	// A clip placed at 15s (timeline range 15-25s for a 10s clip).
	// SubRange(10s, 20s): range overlaps the clip from 15-20s.
	// Visible clip-local window: 0-5s.
	// Shifted StartAt: 15-10 = 5s.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("clip.mp4", 10*time.Second), At(15*time.Second))

	sub := tl.SubRange(10*time.Second, 20*time.Second)
	entries := sub.VideoTracks()[0].Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	// Clip trimmed to first 5s (0-5s clip-local).
	if entries[0].Clip.Duration() != 5*time.Second {
		t.Errorf("expected duration 5s, got %v", entries[0].Clip.Duration())
	}
	// Shifted from 15s to 5s (15 - 10 = 5).
	if entries[0].StartAt != 5*time.Second {
		t.Errorf("expected StartAt 5s, got %v", entries[0].StartAt)
	}
	// Source trim: first 5s of the clip.
	if entries[0].Clip.TrimStart() != 0 {
		t.Errorf("expected TrimStart 0, got %v", entries[0].Clip.TrimStart())
	}
	if entries[0].Clip.TrimEnd() != 5*time.Second {
		t.Errorf("expected TrimEnd 5s, got %v", entries[0].Clip.TrimEnd())
	}
}

func TestSubRange_TextClipInTimeline(t *testing.T) {
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")

	txt := clip.NewText("Chapter 1", clip.DefaultTextStyle()).WithDuration(10 * time.Second)
	track.Add(txt, At(5*time.Second))

	// Range 8-13s: text clip at 5-15s, visible 8-13s → 5s, clip-local 3-8s, starts at 0.
	sub := tl.SubRange(8*time.Second, 13*time.Second)
	entries := sub.VideoTracks()[0].Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Clip.Duration() != 5*time.Second {
		t.Errorf("expected duration 5s, got %v", entries[0].Clip.Duration())
	}
	if entries[0].StartAt != 0 {
		t.Errorf("expected StartAt 0, got %v", entries[0].StartAt)
	}
	// Verify it's still a TextClip with correct content.
	tc, ok := entries[0].Clip.(*clip.TextClip)
	if !ok {
		t.Fatal("expected *TextClip")
	}
	if tc.Text != "Chapter 1" {
		t.Errorf("expected text 'Chapter 1', got %q", tc.Text)
	}
}

func TestSubRange_EffectsPreservedInTimeline(t *testing.T) {
	// Clip with SpeedUp and FadeIn effects should preserve them through timeline SubRange.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")

	v := clip.NewVideoWithDuration("footage.mp4", 30*time.Second).
		WithEffect(effects.SpeedUp(2.0)).
		WithEffect(effects.FadeIn(1 * time.Second))
	track.Add(v, At(0))

	sub := tl.SubRange(0, 10*time.Second)
	entries := sub.VideoTracks()[0].Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	effs := entries[0].Clip.Effects()
	if len(effs) != 2 {
		t.Fatalf("expected 2 effects preserved through timeline SubRange, got %d", len(effs))
	}
	if effs[0].Name() != "speed" {
		t.Errorf("expected first effect 'speed', got %q", effs[0].Name())
	}
	if effs[1].Name() != "fade_in" {
		t.Errorf("expected second effect 'fade_in', got %q", effs[1].Name())
	}
}

func TestDryRunRange_EmptyResult(t *testing.T) {
	// Range entirely in a gap → all clips excluded.
	// DryRun should fail with a validation error (no clips).
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("a.mp4", 10*time.Second), At(0))
	track.Add(clip.NewVideoWithDuration("b.mp4", 10*time.Second), At(20*time.Second))

	_, err := tl.DryRunRange(12*time.Second, 18*time.Second, export.YouTube1080p(), "output.mp4")
	if err == nil {
		t.Error("expected error for DryRunRange with empty result (no clips)")
	}
	if err != nil && !strings.Contains(err.Error(), "compilation failed") {
		t.Errorf("expected compilation failed error, got: %v", err)
	}
}

func TestSubRange_Sequential(t *testing.T) {
	// SubRange of a SubRange: progressive narrowing.
	// 60s clip at 0s. First SubRange(10, 50) → 40s sub.
	// Second SubRange(5, 30) → 25s sub-sub (original source 15-40s).
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("footage.mp4", 60*time.Second), At(0))

	sub1 := tl.SubRange(10*time.Second, 50*time.Second) // 40s timeline, clip source 10-50s
	if sub1.Duration() != 40*time.Second {
		t.Fatalf("sub1: expected 40s duration, got %v", sub1.Duration())
	}

	sub2 := sub1.SubRange(5*time.Second, 30*time.Second) // 25s timeline
	if sub2.Duration() != 25*time.Second {
		t.Fatalf("sub2: expected 25s duration, got %v", sub2.Duration())
	}

	entries := sub2.VideoTracks()[0].Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].StartAt != 0 {
		t.Errorf("expected StartAt 0, got %v", entries[0].StartAt)
	}
	if entries[0].Clip.Duration() != 25*time.Second {
		t.Errorf("expected duration 25s, got %v", entries[0].Clip.Duration())
	}
	// Source coordinates: first SubRange → source 10-50s, clip-local 0-40s.
	// Second SubRange(5, 30) on that clip → source 10 + 5 = 15s, 10 + 30 = 40s.
	if entries[0].Clip.TrimStart() != 15*time.Second {
		t.Errorf("expected TrimStart 15s (double sub-range), got %v", entries[0].Clip.TrimStart())
	}
	if entries[0].Clip.TrimEnd() != 40*time.Second {
		t.Errorf("expected TrimEnd 40s (double sub-range), got %v", entries[0].Clip.TrimEnd())
	}
}

func TestSubRange_SequentialWithDryRun(t *testing.T) {
	// Verify the double SubRange produces a valid FFmpeg command.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("footage.mp4", 60*time.Second), At(0))

	sub := tl.SubRange(10*time.Second, 50*time.Second).SubRange(5*time.Second, 30*time.Second)

	cmd, err := sub.DryRun(export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRun on double SubRange returned error: %v", err)
	}

	cmdStr := cmd.String()
	if !strings.Contains(cmdStr, "footage.mp4") {
		t.Errorf("expected footage.mp4, got:\n%s", cmdStr)
	}
	// Should have trim filter with source coordinates 15-40.
	if !strings.Contains(cmdStr, "start=15.000") {
		t.Errorf("expected trim start=15.000, got:\n%s", cmdStr)
	}
	if !strings.Contains(cmdStr, "end=40.000") {
		t.Errorf("expected trim end=40.000, got:\n%s", cmdStr)
	}
}

func TestSubRange_PreTrimmedClipAtNonZeroPosition(t *testing.T) {
	// Clip trimmed to source 20-40s (20s local), placed at timeline 10s.
	// Timeline range: 10-30s → clip occupies 10-30s, fully visible.
	// SubRange(15, 25): clip visible 15-25s, clip-local 5-15s.
	// Source: 20 + 5 = 25s, 20 + 15 = 35s.
	// Shifted StartAt: max(10-15, 0) = 0.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	v := clip.NewVideoWithDuration("source.mp4", 60*time.Second).
		Trim(20*time.Second, 40*time.Second)
	track.Add(v, At(10*time.Second))

	sub := tl.SubRange(15*time.Second, 25*time.Second)
	entries := sub.VideoTracks()[0].Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].StartAt != 0 {
		t.Errorf("expected StartAt 0, got %v", entries[0].StartAt)
	}
	if entries[0].Clip.Duration() != 10*time.Second {
		t.Errorf("expected duration 10s, got %v", entries[0].Clip.Duration())
	}
	if entries[0].Clip.TrimStart() != 25*time.Second {
		t.Errorf("expected TrimStart 25s (source), got %v", entries[0].Clip.TrimStart())
	}
	if entries[0].Clip.TrimEnd() != 35*time.Second {
		t.Errorf("expected TrimEnd 35s (source), got %v", entries[0].Clip.TrimEnd())
	}
}

func TestDryRunRange_PreTrimmedAtNonZero(t *testing.T) {
	// Verify the FFmpeg command for pre-trimmed clip at non-zero position.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	v := clip.NewVideoWithDuration("source.mp4", 60*time.Second).
		Trim(20*time.Second, 40*time.Second)
	track.Add(v, At(10*time.Second))

	cmd, err := tl.DryRunRange(15*time.Second, 25*time.Second, export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRunRange returned error: %v", err)
	}

	cmdStr := cmd.String()
	if !strings.Contains(cmdStr, "start=25.000") {
		t.Errorf("expected trim start=25.000, got:\n%s", cmdStr)
	}
	if !strings.Contains(cmdStr, "end=35.000") {
		t.Errorf("expected trim end=35.000, got:\n%s", cmdStr)
	}
}

func TestSubRange_WithSpeedEffect(t *testing.T) {
	// 30s clip with SpeedUp(2.0) → Duration()=15s.
	// Placed at 0s. SubRange(5, 10): clip-local 5-10s.
	// Source: ratio=2.0, so source 10-20s. Duration = 5s.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	v := clip.NewVideoWithDuration("footage.mp4", 30*time.Second).
		WithEffect(effects.SpeedUp(2.0))
	track.Add(v, At(0))

	sub := tl.SubRange(5*time.Second, 10*time.Second)
	entries := sub.VideoTracks()[0].Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Clip.Duration() != 5*time.Second {
		t.Errorf("expected duration 5s, got %v", entries[0].Clip.Duration())
	}
	if entries[0].Clip.TrimStart() != 10*time.Second {
		t.Errorf("expected TrimStart 10s (source via ratio), got %v", entries[0].Clip.TrimStart())
	}
	if entries[0].Clip.TrimEnd() != 20*time.Second {
		t.Errorf("expected TrimEnd 20s (source via ratio), got %v", entries[0].Clip.TrimEnd())
	}
	// Effects should be preserved.
	effs := entries[0].Clip.Effects()
	if len(effs) != 1 || effs[0].Name() != "speed" {
		t.Errorf("expected speed effect preserved, got %v", effs)
	}
}

func TestDryRunRange_WithSpeedEffect(t *testing.T) {
	// Verify FFmpeg command for speed-effected sub-ranged clip.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	v := clip.NewVideoWithDuration("footage.mp4", 30*time.Second).
		WithEffect(effects.SpeedUp(2.0)).
		WithEffect(effects.AudioSpeed(2.0))
	track.Add(v, At(0))

	cmd, err := tl.DryRunRange(5*time.Second, 10*time.Second, export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRunRange returned error: %v", err)
	}

	cmdStr := cmd.String()
	// Source coordinates 10-20s.
	if !strings.Contains(cmdStr, "start=10.000") {
		t.Errorf("expected trim start=10.000, got:\n%s", cmdStr)
	}
	if !strings.Contains(cmdStr, "end=20.000") {
		t.Errorf("expected trim end=20.000, got:\n%s", cmdStr)
	}
	// Speed effects should still be applied.
	if !strings.Contains(cmdStr, "setpts") {
		t.Errorf("expected setpts for speed effect, got:\n%s", cmdStr)
	}
	if !strings.Contains(cmdStr, "atempo") {
		t.Errorf("expected atempo for audio speed, got:\n%s", cmdStr)
	}
}
