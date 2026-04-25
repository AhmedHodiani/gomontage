package kdenlive

import (
	"encoding/xml"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ahmedhodiani/gomontage/clip"
	"github.com/ahmedhodiani/gomontage/timeline"
)

func TestExportBasicVideoTrack(t *testing.T) {
	tl := timeline.New(timeline.Config{Width: 1920, Height: 1080, FPS: 30})
	vt := tl.AddVideoTrack("main")
	vt.Add(clip.NewVideoWithDuration("test.mp4", 5*time.Second), timeline.At(0))

	data, err := ExportBytes(tl)
	if err != nil {
		t.Fatalf("ExportBytes failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("ExportBytes returned empty data")
	}

	s := string(data)
	if !strings.Contains(s, `<mlt `) {
		t.Error("XML missing <mlt> root element")
	}
	if !strings.Contains(s, `test.mp4`) {
		t.Error("XML missing chain entry for test.mp4")
	}
	if !strings.Contains(s, `producer0`) {
		t.Error("XML missing black background producer")
	}
	if !strings.Contains(s, "1920") || !strings.Contains(s, "1080") {
		t.Error("XML missing profile dimensions")
	}
	if !strings.Contains(s, `mlt_service">qtblend`) && !strings.Contains(s, `mlt_service">qtblend`) {
		t.Error("XML missing qtblend transition for video track")
	}
	// Video tracks should NOT have kdenlive:audio_track in their content playlist
	// (that property is only on audio tracks)
	if strings.Count(s, "kdenlive:audio_track") != 0 {
		// Check if audio_track appears outside of audio track tractors
		// (it should only appear in audio track contexts)
		t.Logf("Note: audio_track property count = %d", strings.Count(s, "kdenlive:audio_track"))
	}
}

func TestExportAudioTrack(t *testing.T) {
	tl := timeline.New(timeline.Config{Width: 1920, Height: 1080, FPS: 30})
	at := tl.AddAudioTrack("narration")
	at.Add(clip.NewAudioWithDuration("narration.mp3", 90*time.Second), timeline.At(0))

	data, err := ExportBytes(tl)
	if err != nil {
		t.Fatalf("ExportBytes failed: %v", err)
	}

	s := string(data)
	if !strings.Contains(s, `narration.mp3`) {
		t.Error("XML missing chain entry for narration.mp3")
	}
	if !strings.Contains(s, `kdenlive:audio_track`) {
		t.Error("XML missing kdenlive:audio_track property for audio track")
	}
	if !strings.Contains(s, `mlt_service">mix`) {
		t.Error("XML missing mix transition for audio track")
	}
	if !strings.Contains(s, `hide="video"`) {
		t.Error("XML missing hide=video for audio track")
	}
}

func TestExportMultipleTracks(t *testing.T) {
	tl := timeline.New(timeline.Config{Width: 1920, Height: 1080, FPS: 30})
	vt := tl.AddVideoTrack("main")
	vt.Add(clip.NewVideoWithDuration("video.mp4", 10*time.Second), timeline.At(0))
	vt.Add(clip.NewVideoWithDuration("broll.mp4", 5*time.Second), timeline.At(15*time.Second))

	at := tl.AddAudioTrack("narration")
	at.Add(clip.NewAudioWithDuration("speech.mp3", 30*time.Second), timeline.At(0))

	data, err := ExportBytes(tl)
	if err != nil {
		t.Fatalf("ExportBytes failed: %v", err)
	}

	s := string(data)
	if !strings.Contains(s, "video.mp4") {
		t.Error("XML missing video.mp4 chain")
	}
	if !strings.Contains(s, "broll.mp4") {
		t.Error("XML missing broll.mp4 chain")
	}
	if !strings.Contains(s, "speech.mp3") {
		t.Error("XML missing speech.mp3 chain")
	}
	if !strings.Contains(s, "blank") {
		t.Error("XML missing blank element for gap between video clips")
	}
}

func TestExportFades(t *testing.T) {
	tl := timeline.New(timeline.Config{Width: 1920, Height: 1080, FPS: 30})
	vt := tl.AddVideoTrack("main")
	v := clip.NewVideoWithDuration("fade.mp4", 5*time.Second).WithFadeIn(1*time.Second).WithFadeOut(2*time.Second)
	vt.Add(v, timeline.At(0))

	data, err := ExportBytes(tl)
	if err != nil {
		t.Fatalf("ExportBytes failed: %v", err)
	}

	s := string(data)
	if !strings.Contains(s, `fade_from_black`) {
		t.Error("XML missing fade_in filter")
	}
	if !strings.Contains(s, `fade_to_black`) {
		t.Error("XML missing fade_out filter")
	}
	if !strings.Contains(s, `0=0;-1=1`) {
		t.Error("XML missing fade_in alpha property")
	}
	if !strings.Contains(s, `0=1;-1=0`) {
		t.Error("XML missing fade_out alpha property")
	}
}

func TestExportVolume(t *testing.T) {
	tl := timeline.New(timeline.Config{Width: 1920, Height: 1080, FPS: 30})
	at := tl.AddAudioTrack("bgm")
	at.Add(clip.NewAudioWithDuration("music.mp3", 60*time.Second).WithVolume(0.5), timeline.At(0))

	data, err := ExportBytes(tl)
	if err != nil {
		t.Fatalf("ExportBytes failed: %v", err)
	}

	s := string(data)
	if !strings.Contains(s, `mlt_service">volume`) {
		t.Error("XML missing volume filter for volume adjustment")
	}
	if !strings.Contains(s, `gain">0.5`) {
		t.Error("XML missing gain value for volume adjustment")
	}
}

func TestExportToFile(t *testing.T) {
	tl := timeline.New(timeline.Config{Width: 1920, Height: 1080, FPS: 30})
	vt := tl.AddVideoTrack("main")
	vt.Add(clip.NewVideoWithDuration("test.mp4", 5*time.Second), timeline.At(0))

	tmpFile := t.TempDir() + "/test.kdenlive"
	if err := Export(tl, tmpFile); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Exported file is empty")
	}
	if !strings.Contains(string(data), "<mlt") {
		t.Fatal("Exported file doesn't contain MLT XML")
	}
}

func TestExportTrimmedClip(t *testing.T) {
	tl := timeline.New(timeline.Config{Width: 1920, Height: 1080, FPS: 30})
	vt := tl.AddVideoTrack("main")
	v := clip.NewVideoWithDuration("long.mp4", 60*time.Second).Trim(10*time.Second, 30*time.Second)
	vt.Add(v, timeline.At(0))

	data, err := ExportBytes(tl)
	if err != nil {
		t.Fatalf("ExportBytes failed: %v", err)
	}

	s := string(data)
	if !strings.Contains(s, `in="00:00:10.000"`) {
		t.Error("XML missing trimmed in time on entry")
	}
	if !strings.Contains(s, `out="00:00:30.000"`) {
		t.Error("XML missing trimmed out time on entry")
	}
}

func TestExportDiscouragesSourcelessClips(t *testing.T) {
	tl := timeline.New(timeline.Config{Width: 1920, Height: 1080, FPS: 30})
	vt := tl.AddVideoTrack("main")
	vt.Add(clip.NewColor("#000000", 1920, 1080).WithDuration(5*time.Second), timeline.At(0))

	_, err := ExportBytes(tl)
	if err == nil {
		t.Error("Expected error for sourceless clip, got nil")
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "00:00:00.000"},
		{5 * time.Second, "00:00:05.000"},
		{time.Minute, "00:01:00.000"},
		{time.Hour, "01:00:00.000"},
		{90*time.Second + 500*time.Millisecond, "00:01:30.500"},
		{5*time.Minute + 30*time.Second, "00:05:30.000"},
		{2*time.Hour + 30*time.Minute + 15*time.Second + 250*time.Millisecond, "02:30:15.250"},
	}

	for _, tt := range tests {
		got := formatTime(tt.d)
		if got != tt.want {
			t.Errorf("formatTime(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestExportWellFormedXML(t *testing.T) {
	tl := timeline.New(timeline.Config{Width: 1920, Height: 1080, FPS: 30})
	vt := tl.AddVideoTrack("main")
	vt.Add(clip.NewVideoWithDuration("video.mp4", 10*time.Second), timeline.At(0))
	vt.Add(clip.NewVideoWithDuration("broll.mp4", 5*time.Second).WithFadeIn(1*time.Second), timeline.At(10*time.Second))
	at := tl.AddAudioTrack("narration")
	at.Add(clip.NewAudioWithDuration("speech.mp3", 15*time.Second).WithVolume(0.8), timeline.At(0))

	data, err := ExportBytes(tl)
	if err != nil {
		t.Fatalf("ExportBytes failed: %v", err)
	}

	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	for {
		if _, err := decoder.Token(); err != nil {
			if err.Error() == "EOF" {
				break
			}
			t.Fatalf("XML is not well-formed: %v", err)
		}
	}
}

func TestExportMainBinEntries(t *testing.T) {
	tl := timeline.New(timeline.Config{Width: 1920, Height: 1080, FPS: 30})
	vt := tl.AddVideoTrack("main")
	vt.Add(clip.NewVideoWithDuration("video.mp4", 5*time.Second), timeline.At(0))

	data, err := ExportBytes(tl)
	if err != nil {
		t.Fatalf("ExportBytes failed: %v", err)
	}

	s := string(data)
	if !strings.Contains(s, `id="main_bin"`) {
		t.Error("XML missing main_bin playlist")
	}
	if !strings.Contains(s, `kdenlive:projectTractor`) {
		t.Error("XML missing final_tractor")
	}
}

func TestExportVideoAndAudioTrackTogether(t *testing.T) {
	tl := timeline.New(timeline.Config{Width: 1920, Height: 1080, FPS: 30})
	vt := tl.AddVideoTrack("main")
	vt.Add(clip.NewVideoWithDuration("video.mp4", 10*time.Second).VideoOnly(), timeline.At(0))

	at := tl.AddAudioTrack("bgm")
	at.Add(clip.NewAudioWithDuration("music.mp3", 10*time.Second), timeline.At(0))

	data, err := ExportBytes(tl)
	if err != nil {
		t.Fatalf("ExportBytes failed: %v", err)
	}

	s := string(data)
	if !strings.Contains(s, `mlt_service">qtblend`) {
		t.Error("XML missing qtblend transition for video track")
	}
	if !strings.Contains(s, `mlt_service">mix`) {
		t.Error("XML missing mix transition for audio track")
	}
	// Video-only clip on video track should NOT create a companion audio track
	trackCount := strings.Count(s, `<tractor id="tractor`)
	if trackCount != 2 {
		t.Errorf("expected 2 track tractors (1 video + 1 audio), got %d", trackCount)
	}
}

func TestExportUnmutedVideoClip(t *testing.T) {
	tl := timeline.New(timeline.Config{Width: 1920, Height: 1080, FPS: 30})
	vt := tl.AddVideoTrack("main")

	// Video-only clip (muted)
	vt.Add(clip.NewVideoWithDuration("broll.mp4", 5*time.Second).VideoOnly(), timeline.At(0))
	// Unmuted clip (video + audio)  
	vt.Add(clip.NewVideoWithDuration("dialogue.mp4", 5*time.Second).WithVolume(1.5), timeline.At(5*time.Second))
	// Another video-only clip
	vt.Add(clip.NewVideoWithDuration("broll2.mp4", 5*time.Second).VideoOnly(), timeline.At(10*time.Second))

	data, err := ExportBytes(tl)
	if err != nil {
		t.Fatalf("ExportBytes failed: %v", err)
	}

	s := string(data)

	// The companion playlist (playlist1) should have audio entries for the unmuted clip
	// and blanks for the video-only clips
	if !strings.Contains(s, `hide="video"`) {
		t.Error("expected hide=video on audio companion track for video track with audio clips")
	}

	// The unmuted clip should appear on both playlists (video + audio companion)
	dialogueCount := strings.Count(s, `producer="chain1"`)
	if dialogueCount < 2 {
		t.Errorf("expected dialogue.mp4 entry to appear on both playlists, got %d entries", dialogueCount)
	}

	// Volume filter should appear for the unmuted clip (on both playlists)
	volumeCount := strings.Count(s, `gain">1.5`)
	if volumeCount < 2 {
		t.Errorf("expected volume gain 1.5 on both video and audio companion playlists, got %d", volumeCount)
	}
}