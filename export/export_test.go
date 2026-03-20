package export

import (
	"testing"
)

func TestYouTube1080p(t *testing.T) {
	p := YouTube1080p()
	if p.Name != "YouTube 1080p" {
		t.Errorf("expected name 'YouTube 1080p', got %q", p.Name)
	}
	if p.VideoCodec != "libx264" {
		t.Errorf("expected libx264, got %q", p.VideoCodec)
	}
	if p.Width != 1920 || p.Height != 1080 {
		t.Errorf("expected 1920x1080, got %dx%d", p.Width, p.Height)
	}
	if p.AudioCodec != "aac" {
		t.Errorf("expected aac, got %q", p.AudioCodec)
	}

	params := p.Params()
	if params["-c:v"] != "libx264" {
		t.Errorf("expected -c:v libx264, got %q", params["-c:v"])
	}
	if params["-movflags"] != "+faststart" {
		t.Errorf("expected +faststart, got %q", params["-movflags"])
	}
}

func TestReel(t *testing.T) {
	p := Reel()
	if p.Width != 1080 || p.Height != 1920 {
		t.Errorf("expected 1080x1920, got %dx%d", p.Width, p.Height)
	}
}

func TestProRes(t *testing.T) {
	p := ProRes()
	if p.VideoCodec != "prores_ks" {
		t.Errorf("expected prores_ks, got %q", p.VideoCodec)
	}
	if p.Format != "mov" {
		t.Errorf("expected mov, got %q", p.Format)
	}
	params := p.Params()
	if params["-profile:v"] != "2" {
		t.Errorf("expected profile 2, got %q", params["-profile:v"])
	}
	// ProRes should not have CRF.
	if _, ok := params["-crf"]; ok {
		t.Error("ProRes should not have CRF")
	}
}

func TestMP3(t *testing.T) {
	p := MP3()
	if p.AudioCodec != "libmp3lame" {
		t.Errorf("expected libmp3lame, got %q", p.AudioCodec)
	}
	if p.VideoCodec != "" {
		t.Error("MP3 should not have video codec")
	}
	if p.Format != "mp3" {
		t.Errorf("expected mp3, got %q", p.Format)
	}
}

func TestWAV(t *testing.T) {
	p := WAV()
	if p.AudioCodec != "pcm_s16le" {
		t.Errorf("expected pcm_s16le, got %q", p.AudioCodec)
	}
	if p.Format != "wav" {
		t.Errorf("expected wav, got %q", p.Format)
	}
}

func TestProfileBuilder(t *testing.T) {
	p := NewProfile().
		WithName("Custom").
		WithCodec("libx265").
		WithCRF(22).
		WithPreset("slow").
		WithPixelFormat("yuv420p").
		WithResolution(3840, 2160).
		WithMaxRate("20M").
		WithBufSize("40M").
		WithAudioCodec("aac").
		WithAudioBitrate("256k").
		WithAudioSampleRate(48000).
		WithAudioChannels(2).
		WithFormat("mp4").
		WithFastStart().
		WithExtra("-threads", "4").
		Build()

	if p.Name != "Custom" {
		t.Errorf("expected Custom, got %q", p.Name)
	}
	if p.VideoCodec != "libx265" {
		t.Errorf("expected libx265, got %q", p.VideoCodec)
	}
	if p.CRF != 22 {
		t.Errorf("expected CRF 22, got %d", p.CRF)
	}
	if p.Width != 3840 || p.Height != 2160 {
		t.Errorf("expected 3840x2160, got %dx%d", p.Width, p.Height)
	}

	params := p.Params()
	if params["-c:v"] != "libx265" {
		t.Errorf("expected -c:v libx265, got %q", params["-c:v"])
	}
	if params["-threads"] != "4" {
		t.Errorf("expected -threads 4, got %q", params["-threads"])
	}
	if params["-movflags"] != "+faststart" {
		t.Errorf("expected +faststart, got %q", params["-movflags"])
	}
}

func TestProfile_ParamsOmitsUnset(t *testing.T) {
	p := &Profile{
		VideoCodec: "libx264",
		CRF:        -1, // Unset.
	}
	params := p.Params()

	if _, ok := params["-crf"]; ok {
		t.Error("CRF=-1 should not produce -crf param")
	}
	if _, ok := params["-b:v"]; ok {
		t.Error("empty bitrate should not produce -b:v param")
	}
	if _, ok := params["-c:a"]; ok {
		t.Error("empty audio codec should not produce -c:a param")
	}
}

func TestAllPresets(t *testing.T) {
	presets := []*Profile{
		YouTube1080p(),
		YouTube4K(),
		YouTube720p(),
		Reel(),
		ProRes(),
		H265(),
		GIF(),
		MP3(),
		WAV(),
		FLAC(),
	}

	for _, p := range presets {
		if p.Name == "" {
			t.Error("preset has empty name")
		}
		params := p.Params()
		if len(params) == 0 {
			t.Errorf("preset %q has no params", p.Name)
		}
	}
}
