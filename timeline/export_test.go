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

func TestDryRun_TrimFromZero(t *testing.T) {
	// Regression test: Trim(0, 10s) on a 60s clip must produce a trim filter.
	// Previously the compiler's condition missed this case, outputting the full clip.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")

	video := clip.NewVideoWithDuration("long.mp4", 60*time.Second)
	trimmed := video.Trim(0, 10*time.Second)
	track.Add(trimmed, At(0))

	cmd, err := tl.DryRun(export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRun returned error: %v", err)
	}

	cmdStr := cmd.String()

	// The trim filter must be present.
	if !strings.Contains(cmdStr, "trim") {
		t.Errorf("expected trim filter in command for Trim(0, 10s), got:\n%s", cmdStr)
	}

	// PTS reset must follow trim.
	if !strings.Contains(cmdStr, "setpts") {
		t.Errorf("expected setpts filter after trim, got:\n%s", cmdStr)
	}
}

func TestDryRun_TrimFromMiddle(t *testing.T) {
	// Trim from the middle of a clip should also produce trim filter.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")

	video := clip.NewVideoWithDuration("long.mp4", 60*time.Second)
	trimmed := video.Trim(20*time.Second, 40*time.Second)
	track.Add(trimmed, At(0))

	cmd, err := tl.DryRun(export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRun returned error: %v", err)
	}

	cmdStr := cmd.String()

	if !strings.Contains(cmdStr, "trim") {
		t.Errorf("expected trim filter in command for Trim(20s, 40s), got:\n%s", cmdStr)
	}
}

func TestDryRun_UntrimmedClip_NoTrimFilter(t *testing.T) {
	// An untrimmed clip should NOT have a trim filter.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")
	track.Add(clip.NewVideoWithDuration("full.mp4", 10*time.Second), At(0))

	cmd, err := tl.DryRun(export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRun returned error: %v", err)
	}

	cmdStr := cmd.String()

	// There should be no trim filter for an untrimmed clip.
	if strings.Contains(cmdStr, "=start=") {
		t.Errorf("untrimmed clip should not have trim filter, got:\n%s", cmdStr)
	}
}

func TestDryRun_TrimmedAudioFromVideo(t *testing.T) {
	// When a video clip with audio is trimmed, the audio stream should also be trimmed.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")

	video := clip.NewVideoWithDuration("interview.mp4", 60*time.Second)
	trimmed := video.Trim(5*time.Second, 15*time.Second)
	track.Add(trimmed, At(0))

	cmd, err := tl.DryRun(export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRun returned error: %v", err)
	}

	cmdStr := cmd.String()

	// Both video trim and audio trim should be present.
	if !strings.Contains(cmdStr, "trim") {
		t.Errorf("expected trim filter, got:\n%s", cmdStr)
	}
	if !strings.Contains(cmdStr, "atrim") {
		t.Errorf("expected atrim filter for audio from trimmed video, got:\n%s", cmdStr)
	}
}

func TestClip_IsTrimmed(t *testing.T) {
	// Untrimmed clip should report IsTrimmed=false.
	v := clip.NewVideoWithDuration("test.mp4", 60*time.Second)
	if v.IsTrimmed() {
		t.Error("untrimmed video clip should have IsTrimmed()=false")
	}

	// Trimmed clip should report IsTrimmed=true.
	trimmed := v.Trim(0, 10*time.Second)
	if !trimmed.IsTrimmed() {
		t.Error("trimmed video clip should have IsTrimmed()=true")
	}

	// Audio clips too.
	a := clip.NewAudioWithDuration("test.wav", 60*time.Second)
	if a.IsTrimmed() {
		t.Error("untrimmed audio clip should have IsTrimmed()=false")
	}

	aTrimmed := a.Trim(10*time.Second, 30*time.Second)
	if !aTrimmed.IsTrimmed() {
		t.Error("trimmed audio clip should have IsTrimmed()=true")
	}
}

func TestDryRun_ConcatAudioFromVideo(t *testing.T) {
	// Regression test: when multiple trimmed video clips (with audio) are placed
	// on a video track, the audio must be concatenated in sync with the video,
	// NOT mixed with amix (which would overlap both audio streams at time 0).
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})

	video := clip.NewVideoWithDuration("s1e1.mp4", 60*time.Second)
	intro := video.Trim(5*time.Second, 10*time.Second)
	outro := video.Trim(15*time.Second, 20*time.Second)

	track := tl.AddVideoTrack("main")
	track.AddSequence(intro, outro)

	cmd, err := tl.DryRun(export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRun returned error: %v", err)
	}

	cmdStr := cmd.String()

	// The concat filter must include audio (a=1), not video-only (a=0).
	if !strings.Contains(cmdStr, "a=1") {
		t.Errorf("expected concat with a=1 for audio, got:\n%s", cmdStr)
	}

	// There should be NO amix — audio from video clips should be concatenated, not mixed.
	if strings.Contains(cmdStr, "amix") {
		t.Errorf("audio from video track clips should be concatenated, not mixed with amix, got:\n%s", cmdStr)
	}

	// Both atrim filters should be present (one per clip).
	if !strings.Contains(cmdStr, "atrim") {
		t.Errorf("expected atrim filters for audio trimming, got:\n%s", cmdStr)
	}
}

func TestDryRun_ConcatVideoWithBGMusic(t *testing.T) {
	// When video clips are concatenated AND there's a separate audio track,
	// the video audio should be concatenated and then mixed with the bg music.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})

	video := clip.NewVideoWithDuration("footage.mp4", 60*time.Second)
	part1 := video.Trim(0, 10*time.Second)
	part2 := video.Trim(20*time.Second, 30*time.Second)

	vTrack := tl.AddVideoTrack("main")
	vTrack.AddSequence(part1, part2)

	aTrack := tl.AddAudioTrack("music")
	bgm := clip.NewAudioWithDuration("bgm.mp3", 30*time.Second)
	aTrack.Add(bgm, At(0))

	cmd, err := tl.DryRun(export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRun returned error: %v", err)
	}

	cmdStr := cmd.String()

	// Should have concat with audio (for video clips).
	if !strings.Contains(cmdStr, "a=1") {
		t.Errorf("expected concat with a=1, got:\n%s", cmdStr)
	}

	// Should have amix to layer the concatenated audio with the bg music.
	if !strings.Contains(cmdStr, "amix") {
		t.Errorf("expected amix to mix video audio with bg music, got:\n%s", cmdStr)
	}
}

func TestDryRun_VideoOnlyClipsConcat(t *testing.T) {
	// When video clips have no audio (VideoOnly), concat should use a=0.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})

	video := clip.NewVideoWithDuration("footage.mp4", 60*time.Second)
	part1 := video.Trim(0, 10*time.Second).VideoOnly()
	part2 := video.Trim(20*time.Second, 30*time.Second).VideoOnly()

	track := tl.AddVideoTrack("main")
	track.AddSequence(part1, part2)

	cmd, err := tl.DryRun(export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRun returned error: %v", err)
	}

	cmdStr := cmd.String()

	// With VideoOnly clips, concat should be video-only (a=0).
	if strings.Contains(cmdStr, "a=1") {
		t.Errorf("VideoOnly clips should produce concat with a=0, got:\n%s", cmdStr)
	}
	if strings.Contains(cmdStr, "atrim") {
		t.Errorf("VideoOnly clips should not have atrim filters, got:\n%s", cmdStr)
	}
}

func TestDryRun_VideoStartAt(t *testing.T) {
	// Regression test: a single video clip placed at 50s should produce a tpad
	// filter to delay the video by 50 seconds of black frames.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})

	video := clip.NewVideoWithDuration("s1e1.mp4", 60*time.Second)
	intro := video.Trim(5*time.Second, 10*time.Second)

	track := tl.AddVideoTrack("main")
	track.Add(intro, At(50*time.Second))

	cmd, err := tl.DryRun(export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRun returned error: %v", err)
	}

	cmdStr := cmd.String()

	// The tpad filter must be present to delay the video.
	if !strings.Contains(cmdStr, "tpad") {
		t.Errorf("expected tpad filter for video placed at 50s, got:\n%s", cmdStr)
	}
	if !strings.Contains(cmdStr, "start_duration=50.000") {
		t.Errorf("expected tpad start_duration=50.000, got:\n%s", cmdStr)
	}
}

func TestDryRun_VideoStartAt_WithAudio(t *testing.T) {
	// When a video clip with audio is placed at a non-zero time, both the video
	// and its audio must be delayed — tpad for video, adelay for audio.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})

	video := clip.NewVideoWithDuration("clip.mp4", 60*time.Second)
	segment := video.Trim(10*time.Second, 20*time.Second)

	track := tl.AddVideoTrack("main")
	track.Add(segment, At(30*time.Second))

	cmd, err := tl.DryRun(export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRun returned error: %v", err)
	}

	cmdStr := cmd.String()

	// Video must be delayed with tpad.
	if !strings.Contains(cmdStr, "tpad") {
		t.Errorf("expected tpad for video delay, got:\n%s", cmdStr)
	}
	if !strings.Contains(cmdStr, "start_duration=30.000") {
		t.Errorf("expected tpad start_duration=30.000, got:\n%s", cmdStr)
	}

	// Audio must be delayed with adelay (30s = 30000ms).
	if !strings.Contains(cmdStr, "adelay") {
		t.Errorf("expected adelay for audio from video clip, got:\n%s", cmdStr)
	}
	if !strings.Contains(cmdStr, "30000") {
		t.Errorf("expected adelay of 30000ms, got:\n%s", cmdStr)
	}
}

func TestDryRun_VideoStartAtZero_NoTpad(t *testing.T) {
	// A clip placed at At(0) should NOT produce a tpad filter.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})

	video := clip.NewVideoWithDuration("clip.mp4", 10*time.Second)
	track := tl.AddVideoTrack("main")
	track.Add(video, At(0))

	cmd, err := tl.DryRun(export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRun returned error: %v", err)
	}

	cmdStr := cmd.String()
	if strings.Contains(cmdStr, "tpad") {
		t.Errorf("clip at time 0 should not have tpad, got:\n%s", cmdStr)
	}
}
