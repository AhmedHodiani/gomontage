package export

import "fmt"

// Profile encapsulates all FFmpeg output settings for a video export.
// It translates into command-line flags passed to FFmpeg.
type Profile struct {
	// Name is a human-readable label for this profile (e.g. "YouTube 1080p").
	Name string

	// VideoCodec is the FFmpeg video codec (e.g. "libx264", "libx265", "prores_ks").
	// Empty means no video output (audio-only).
	VideoCodec string

	// VideoBitrate is the target video bitrate (e.g. "5M", "8000k").
	// Empty means let FFmpeg decide (CRF-based).
	VideoBitrate string

	// CRF is the Constant Rate Factor for quality-based encoding (0-51 for H.264).
	// -1 means unset (use bitrate instead).
	CRF int

	// Preset is the encoding speed/quality tradeoff (e.g. "medium", "slow", "fast").
	Preset string

	// PixelFormat is the pixel format (e.g. "yuv420p", "yuv422p10le").
	PixelFormat string

	// Width is the output width. 0 means use timeline width.
	Width int

	// Height is the output height. 0 means use timeline height.
	Height int

	// MaxRate limits the maximum bitrate for VBV buffering.
	MaxRate string

	// BufSize is the VBV buffer size.
	BufSize string

	// AudioCodec is the FFmpeg audio codec (e.g. "aac", "libmp3lame", "pcm_s16le").
	// Empty means no audio output.
	AudioCodec string

	// AudioBitrate is the target audio bitrate (e.g. "192k", "320k").
	AudioBitrate string

	// AudioSampleRate is the output sample rate in Hz (e.g. 44100, 48000).
	// 0 means use source sample rate.
	AudioSampleRate int

	// AudioChannels is the number of output channels. 0 means use source channels.
	AudioChannels int

	// Format is the container format (e.g. "mp4", "mov", "mkv", "mp3").
	// Usually inferred from the output file extension.
	Format string

	// MovFlags are additional flags for MP4/MOV containers (e.g. "+faststart").
	MovFlags string

	// ExtraArgs holds any additional FFmpeg arguments not covered above.
	ExtraArgs map[string]string
}

// Params converts the profile into a map of FFmpeg output parameters.
// This is used by the timeline compiler to set output options.
func (p *Profile) Params() map[string]string {
	params := make(map[string]string)

	if p.VideoCodec != "" {
		params["-c:v"] = p.VideoCodec
	}
	if p.VideoBitrate != "" {
		params["-b:v"] = p.VideoBitrate
	}
	if p.CRF >= 0 {
		params["-crf"] = intToStr(p.CRF)
	}
	if p.Preset != "" {
		params["-preset"] = p.Preset
	}
	if p.PixelFormat != "" {
		params["-pix_fmt"] = p.PixelFormat
	}
	if p.MaxRate != "" {
		params["-maxrate"] = p.MaxRate
	}
	if p.BufSize != "" {
		params["-bufsize"] = p.BufSize
	}
	if p.AudioCodec != "" {
		params["-c:a"] = p.AudioCodec
	}
	if p.AudioBitrate != "" {
		params["-b:a"] = p.AudioBitrate
	}
	if p.AudioSampleRate > 0 {
		params["-ar"] = intToStr(p.AudioSampleRate)
	}
	if p.AudioChannels > 0 {
		params["-ac"] = intToStr(p.AudioChannels)
	}
	if p.Format != "" {
		params["-f"] = p.Format
	}
	if p.MovFlags != "" {
		params["-movflags"] = p.MovFlags
	}
	for k, v := range p.ExtraArgs {
		params[k] = v
	}

	return params
}

func intToStr(n int) string {
	return fmt.Sprintf("%d", n)
}
