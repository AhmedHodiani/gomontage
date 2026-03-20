package export

// YouTube1080p returns a profile optimized for YouTube at 1080p.
// H.264 + AAC with faststart for web streaming.
func YouTube1080p() *Profile {
	return &Profile{
		Name:            "YouTube 1080p",
		VideoCodec:      "libx264",
		CRF:             18,
		Preset:          "medium",
		PixelFormat:     "yuv420p",
		Width:           1920,
		Height:          1080,
		MaxRate:         "10M",
		BufSize:         "20M",
		AudioCodec:      "aac",
		AudioBitrate:    "192k",
		AudioSampleRate: 48000,
		AudioChannels:   2,
		MovFlags:        "+faststart",
	}
}

// YouTube4K returns a profile optimized for YouTube at 4K.
func YouTube4K() *Profile {
	return &Profile{
		Name:            "YouTube 4K",
		VideoCodec:      "libx264",
		CRF:             18,
		Preset:          "medium",
		PixelFormat:     "yuv420p",
		Width:           3840,
		Height:          2160,
		MaxRate:         "40M",
		BufSize:         "80M",
		AudioCodec:      "aac",
		AudioBitrate:    "192k",
		AudioSampleRate: 48000,
		AudioChannels:   2,
		MovFlags:        "+faststart",
	}
}

// YouTube720p returns a profile optimized for YouTube at 720p.
func YouTube720p() *Profile {
	return &Profile{
		Name:            "YouTube 720p",
		VideoCodec:      "libx264",
		CRF:             20,
		Preset:          "medium",
		PixelFormat:     "yuv420p",
		Width:           1280,
		Height:          720,
		MaxRate:         "5M",
		BufSize:         "10M",
		AudioCodec:      "aac",
		AudioBitrate:    "128k",
		AudioSampleRate: 48000,
		AudioChannels:   2,
		MovFlags:        "+faststart",
	}
}

// Reel returns a profile for vertical short-form video (Instagram Reels, TikTok, YouTube Shorts).
// 1080x1920 (9:16 aspect ratio).
func Reel() *Profile {
	return &Profile{
		Name:            "Reel (9:16)",
		VideoCodec:      "libx264",
		CRF:             18,
		Preset:          "medium",
		PixelFormat:     "yuv420p",
		Width:           1080,
		Height:          1920,
		MaxRate:         "10M",
		BufSize:         "20M",
		AudioCodec:      "aac",
		AudioBitrate:    "192k",
		AudioSampleRate: 48000,
		AudioChannels:   2,
		MovFlags:        "+faststart",
	}
}

// ProRes returns a profile for Apple ProRes 422 — a professional intermediate codec
// commonly used for editing and color grading workflows.
func ProRes() *Profile {
	return &Profile{
		Name:        "ProRes 422",
		VideoCodec:  "prores_ks",
		PixelFormat: "yuv422p10le",
		CRF:         -1, // ProRes doesn't use CRF.
		AudioCodec:  "pcm_s16le",
		Format:      "mov",
		ExtraArgs: map[string]string{
			"-profile:v": "2", // ProRes 422 Normal.
		},
	}
}

// H265 returns a profile using H.265/HEVC for better compression at the same quality.
func H265() *Profile {
	return &Profile{
		Name:            "H.265/HEVC",
		VideoCodec:      "libx265",
		CRF:             22,
		Preset:          "medium",
		PixelFormat:     "yuv420p",
		AudioCodec:      "aac",
		AudioBitrate:    "192k",
		AudioSampleRate: 48000,
		AudioChannels:   2,
		MovFlags:        "+faststart",
	}
}

// GIF returns a profile for animated GIF output.
// Note: GIFs have no audio.
func GIF() *Profile {
	return &Profile{
		Name:   "Animated GIF",
		Format: "gif",
		CRF:    -1,
	}
}

// MP3 returns an audio-only MP3 export profile.
func MP3() *Profile {
	return &Profile{
		Name:            "MP3",
		AudioCodec:      "libmp3lame",
		AudioBitrate:    "192k",
		AudioSampleRate: 44100,
		AudioChannels:   2,
		Format:          "mp3",
		CRF:             -1,
	}
}

// WAV returns an audio-only WAV export profile (uncompressed PCM).
func WAV() *Profile {
	return &Profile{
		Name:            "WAV",
		AudioCodec:      "pcm_s16le",
		AudioSampleRate: 48000,
		AudioChannels:   2,
		Format:          "wav",
		CRF:             -1,
	}
}

// FLAC returns an audio-only FLAC export profile (lossless compressed).
func FLAC() *Profile {
	return &Profile{
		Name:            "FLAC",
		AudioCodec:      "flac",
		AudioSampleRate: 48000,
		AudioChannels:   2,
		Format:          "flac",
		CRF:             -1,
	}
}
