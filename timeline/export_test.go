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

func TestDryRun_SequenceNoGap(t *testing.T) {
	// Regression test: AddSequence places clips back-to-back by setting
	// incremental StartAt values. The compiler must NOT apply tpad/adelay to
	// individual clips in a sequence — concat handles ordering. Previously,
	// tpad was applied per-clip, prepending black frames before the second clip
	// and producing an unexpected gap.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})

	video := clip.NewVideoWithDuration("s1e1.mp4", 60*time.Second)
	intro := video.Trim(5*time.Second, 10*time.Second)  // 5s
	outro := video.Trim(10*time.Second, 25*time.Second) // 15s

	track := tl.AddVideoTrack("main")
	track.AddSequence(intro, outro)

	cmd, err := tl.DryRun(export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRun returned error: %v", err)
	}

	cmdStr := cmd.String()

	// The concat filter must be present.
	if !strings.Contains(cmdStr, "concat") {
		t.Errorf("expected concat filter for sequence, got:\n%s", cmdStr)
	}

	// There must be NO tpad — individual clips in a sequence should not be padded.
	if strings.Contains(cmdStr, "tpad") {
		t.Errorf("sequence clips should not have tpad (causes black gap), got:\n%s", cmdStr)
	}

	// There must be NO adelay on per-clip audio — audio is concatenated, not delayed.
	if strings.Contains(cmdStr, "adelay") {
		t.Errorf("sequence clips should not have adelay (audio is concatenated), got:\n%s", cmdStr)
	}
}

func TestDryRun_SequenceWithInitialOffset(t *testing.T) {
	// When multiple clips start at a non-zero time (e.g., first clip at 10s),
	// the compiler should insert a black gap clip before the first clip instead
	// of using tpad on the concat output. This keeps all timing within the
	// concat filter graph.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})

	video := clip.NewVideoWithDuration("footage.mp4", 60*time.Second)
	part1 := video.Trim(0, 5*time.Second)               // 5s
	part2 := video.Trim(10*time.Second, 20*time.Second) // 10s

	track := tl.AddVideoTrack("main")
	// Manually place: first clip at 10s, second at 15s.
	track.Add(part1, At(10*time.Second))
	track.Add(part2, At(15*time.Second))

	cmd, err := tl.DryRun(export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRun returned error: %v", err)
	}

	cmdStr := cmd.String()

	// The concat filter must be present (gap + two clips = n=3).
	if !strings.Contains(cmdStr, "concat") {
		t.Errorf("expected concat filter, got:\n%s", cmdStr)
	}
	if !strings.Contains(cmdStr, "n=3") {
		t.Errorf("expected concat with n=3 (gap + 2 clips), got:\n%s", cmdStr)
	}

	// A black color source should fill the 10s gap before the first clip.
	if !strings.Contains(cmdStr, "color=c=black") {
		t.Errorf("expected black color gap clip for initial 10s offset, got:\n%s", cmdStr)
	}
	if !strings.Contains(cmdStr, "d=10.000") {
		t.Errorf("expected gap duration of 10.000s, got:\n%s", cmdStr)
	}

	// No tpad — multi-clip case uses gap insertion, not tpad.
	if strings.Contains(cmdStr, "tpad") {
		t.Errorf("multi-clip with initial offset should use gap insertion, not tpad, got:\n%s", cmdStr)
	}
}

func TestDryRun_ImageClipInputParams(t *testing.T) {
	// Image clips need -loop 1, -framerate, and -t input-level flags so FFmpeg
	// produces a proper video stream from a still image.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")

	img := clip.NewImage("title.png").WithDuration(4*time.Second).WithSize(1920, 1080)
	track.Add(img, At(0))

	cmd, err := tl.DryRun(export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRun returned error: %v", err)
	}

	cmdStr := cmd.String()

	// Input-level params must appear before -i.
	if !strings.Contains(cmdStr, "-loop 1") {
		t.Errorf("expected -loop 1 for image input, got:\n%s", cmdStr)
	}
	if !strings.Contains(cmdStr, "-framerate 30") {
		t.Errorf("expected -framerate 30 for image input, got:\n%s", cmdStr)
	}
	if !strings.Contains(cmdStr, "-t 4.000") {
		t.Errorf("expected -t 4.000 for image input, got:\n%s", cmdStr)
	}

	// Pixel format normalization must be present for concat compatibility.
	if !strings.Contains(cmdStr, "format=pix_fmts=yuv420p") {
		t.Errorf("expected format=pix_fmts=yuv420p for image clip, got:\n%s", cmdStr)
	}
}

func TestDryRun_ImageThenVideoConcat(t *testing.T) {
	// An image followed by a video clip should produce a concat filter.
	// Both must be scaled and the image must have yuv420p format.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})
	track := tl.AddVideoTrack("main")

	img := clip.NewImage("titlecard.png").WithDuration(4*time.Second).WithSize(1920, 1080)
	video := clip.NewVideoWithDuration("footage.mp4", 30*time.Second).VideoOnly()

	track.AddSequence(img, video)

	cmd, err := tl.DryRun(export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRun returned error: %v", err)
	}

	cmdStr := cmd.String()

	// Concat filter must be present.
	if !strings.Contains(cmdStr, "concat") {
		t.Errorf("expected concat filter for image+video sequence, got:\n%s", cmdStr)
	}

	// Image input must have loop/framerate/duration flags.
	if !strings.Contains(cmdStr, "-loop 1") {
		t.Errorf("expected -loop 1 for image in sequence, got:\n%s", cmdStr)
	}

	// Both inputs must be present.
	if !strings.Contains(cmdStr, "titlecard.png") {
		t.Errorf("expected titlecard.png input, got:\n%s", cmdStr)
	}
	if !strings.Contains(cmdStr, "footage.mp4") {
		t.Errorf("expected footage.mp4 input, got:\n%s", cmdStr)
	}
}

func TestDryRun_ConcatOutputPadOrder(t *testing.T) {
	// Regression test: when concat produces both video and audio outputs (v=1, a=1)
	// and the audio is then mixed with independent audio tracks via amix, the
	// concat output labels must be ordered video-first so FFmpeg maps them to
	// the correct output pads (pad 0 = video, pad 1 = audio). Previously the
	// audio label could appear first, causing a media type mismatch error:
	// "Media type mismatch between concat output pad 0 (video) and amix input pad 0 (audio)"
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})

	video := clip.NewVideoWithDuration("episode.mp4", 600*time.Second)
	// Mix of video-only and with-audio clips — this triggers concat with a=1
	// plus anullsrc silence placeholders for the video-only clips.
	muted := video.Trim(10*time.Second, 15*time.Second).VideoOnly()
	unmuted := video.Trim(20*time.Second, 28*time.Second) // has audio

	vTrack := tl.AddVideoTrack("main")
	vTrack.AddSequence(muted, unmuted)

	// Independent audio track forces the concat audio output through amix.
	aTrack := tl.AddAudioTrack("narration")
	narration := clip.NewAudioWithDuration("voiceover.mp3", 30*time.Second)
	aTrack.Add(narration, At(0))

	cmd, err := tl.DryRun(export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRun returned error: %v", err)
	}

	cmdStr := cmd.String()

	// The concat filter must produce both video and audio.
	if !strings.Contains(cmdStr, "a=1") {
		t.Errorf("expected concat with a=1, got:\n%s", cmdStr)
	}

	// The amix filter must receive the audio output from concat, not the video.
	if !strings.Contains(cmdStr, "amix") {
		t.Errorf("expected amix filter, got:\n%s", cmdStr)
	}

	// Verify output label ordering: in the concat filter output, the vconcat label
	// must appear before the aconcat label so FFmpeg maps video to pad 0.
	concatIdx := strings.Index(cmdStr, "concat=")
	if concatIdx < 0 {
		t.Fatalf("concat filter not found in command")
	}
	afterConcat := cmdStr[concatIdx:]
	vconcatIdx := strings.Index(afterConcat, "[vconcat")
	aconcatIdx := strings.Index(afterConcat, "[aconcat")
	if vconcatIdx < 0 || aconcatIdx < 0 {
		t.Fatalf("expected both vconcat and aconcat labels, got:\n%s", afterConcat)
	}
	if vconcatIdx > aconcatIdx {
		t.Errorf("concat output labels are misordered: video label must come before audio label for correct FFmpeg pad mapping, got:\n%s", afterConcat)
	}
}

func TestDryRun_NonContiguousClipsInsertGaps(t *testing.T) {
	// Regression test for Bug #2: when clips are placed at non-contiguous
	// absolute times via Add(clip, At(T)), the compiler must insert black gap
	// clips so that each clip appears at its correct position in the output.
	// Previously, all clips were concatenated back-to-back, ignoring StartAt.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})

	video := clip.NewVideoWithDuration("footage.mp4", 120*time.Second)
	clipA := video.Trim(0, 5*time.Second).VideoOnly()               // 5s, placed at 0s
	clipB := video.Trim(10*time.Second, 15*time.Second).VideoOnly() // 5s, placed at 20s (15s gap)
	clipC := video.Trim(30*time.Second, 35*time.Second).VideoOnly() // 5s, placed at 40s (15s gap)

	track := tl.AddVideoTrack("main")
	track.Add(clipA, At(0))
	track.Add(clipB, At(20*time.Second))
	track.Add(clipC, At(40*time.Second))

	cmd, err := tl.DryRun(export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRun returned error: %v", err)
	}

	cmdStr := cmd.String()

	// Should have concat with n=5: clipA + gap(15s) + clipB + gap(15s) + clipC.
	if !strings.Contains(cmdStr, "n=5") {
		t.Errorf("expected concat with n=5 (3 clips + 2 gaps), got:\n%s", cmdStr)
	}

	// Should have black color gaps.
	if !strings.Contains(cmdStr, "color=c=black") {
		t.Errorf("expected black gap clips between non-contiguous clips, got:\n%s", cmdStr)
	}

	// Each gap should be 15s.
	if !strings.Contains(cmdStr, "d=15.000") {
		t.Errorf("expected 15s gap durations, got:\n%s", cmdStr)
	}

	// No tpad — gaps handle all positioning.
	if strings.Contains(cmdStr, "tpad") {
		t.Errorf("should not have tpad when gaps handle positioning, got:\n%s", cmdStr)
	}
}

func TestDryRun_AmixNormalize(t *testing.T) {
	// The amix filter must include normalize=0 to prevent FFmpeg from reducing
	// each input's volume by dividing by the number of inputs.
	tl := New(Config{Width: 1920, Height: 1080, FPS: 30})

	video := clip.NewVideoWithDuration("footage.mp4", 30*time.Second)
	vTrack := tl.AddVideoTrack("main")
	vTrack.Add(video, At(0))

	aTrack := tl.AddAudioTrack("narration")
	narration := clip.NewAudioWithDuration("narration.mp3", 30*time.Second)
	aTrack.Add(narration, At(0))

	cmd, err := tl.DryRun(export.YouTube1080p(), "output.mp4")
	if err != nil {
		t.Fatalf("DryRun returned error: %v", err)
	}

	cmdStr := cmd.String()

	// amix must be present (video audio + narration).
	if !strings.Contains(cmdStr, "amix") {
		t.Errorf("expected amix filter, got:\n%s", cmdStr)
	}

	// normalize=0 must be present to preserve volume levels.
	if !strings.Contains(cmdStr, "normalize=0") {
		t.Errorf("expected normalize=0 in amix filter to prevent volume reduction, got:\n%s", cmdStr)
	}
}
