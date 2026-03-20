package export

// ProfileBuilder provides a fluent API for constructing custom export profiles.
//
// Example:
//
//	profile := export.NewProfile().
//	    WithName("My Custom Profile").
//	    WithCodec("libx265").
//	    WithCRF(20).
//	    WithAudioCodec("aac").
//	    WithAudioBitrate("256k").
//	    Build()
type ProfileBuilder struct {
	profile Profile
}

// NewProfile starts building a custom export profile.
func NewProfile() *ProfileBuilder {
	return &ProfileBuilder{
		profile: Profile{
			CRF: -1, // Unset by default.
		},
	}
}

// WithName sets the profile name.
func (b *ProfileBuilder) WithName(name string) *ProfileBuilder {
	b.profile.Name = name
	return b
}

// WithCodec sets the video codec (e.g. "libx264", "libx265", "prores_ks").
func (b *ProfileBuilder) WithCodec(codec string) *ProfileBuilder {
	b.profile.VideoCodec = codec
	return b
}

// WithBitrate sets the target video bitrate (e.g. "5M", "8000k").
func (b *ProfileBuilder) WithBitrate(bitrate string) *ProfileBuilder {
	b.profile.VideoBitrate = bitrate
	return b
}

// WithCRF sets the Constant Rate Factor (0-51 for H.264, 0-63 for H.265).
func (b *ProfileBuilder) WithCRF(crf int) *ProfileBuilder {
	b.profile.CRF = crf
	return b
}

// WithPreset sets the encoding speed/quality tradeoff (e.g. "ultrafast", "medium", "slow").
func (b *ProfileBuilder) WithPreset(preset string) *ProfileBuilder {
	b.profile.Preset = preset
	return b
}

// WithPixelFormat sets the pixel format (e.g. "yuv420p", "yuv444p").
func (b *ProfileBuilder) WithPixelFormat(fmt string) *ProfileBuilder {
	b.profile.PixelFormat = fmt
	return b
}

// WithResolution sets the output resolution.
func (b *ProfileBuilder) WithResolution(width, height int) *ProfileBuilder {
	b.profile.Width = width
	b.profile.Height = height
	return b
}

// WithMaxRate sets the maximum bitrate for VBV buffering.
func (b *ProfileBuilder) WithMaxRate(rate string) *ProfileBuilder {
	b.profile.MaxRate = rate
	return b
}

// WithBufSize sets the VBV buffer size.
func (b *ProfileBuilder) WithBufSize(size string) *ProfileBuilder {
	b.profile.BufSize = size
	return b
}

// WithAudioCodec sets the audio codec (e.g. "aac", "libmp3lame", "pcm_s16le").
func (b *ProfileBuilder) WithAudioCodec(codec string) *ProfileBuilder {
	b.profile.AudioCodec = codec
	return b
}

// WithAudioBitrate sets the target audio bitrate (e.g. "192k", "320k").
func (b *ProfileBuilder) WithAudioBitrate(bitrate string) *ProfileBuilder {
	b.profile.AudioBitrate = bitrate
	return b
}

// WithAudioSampleRate sets the output sample rate in Hz.
func (b *ProfileBuilder) WithAudioSampleRate(rate int) *ProfileBuilder {
	b.profile.AudioSampleRate = rate
	return b
}

// WithAudioChannels sets the number of output audio channels (1=mono, 2=stereo).
func (b *ProfileBuilder) WithAudioChannels(channels int) *ProfileBuilder {
	b.profile.AudioChannels = channels
	return b
}

// WithFormat sets the container format (e.g. "mp4", "mkv", "webm").
func (b *ProfileBuilder) WithFormat(format string) *ProfileBuilder {
	b.profile.Format = format
	return b
}

// WithFastStart enables faststart for MP4/MOV containers (moves moov atom to beginning).
func (b *ProfileBuilder) WithFastStart() *ProfileBuilder {
	b.profile.MovFlags = "+faststart"
	return b
}

// WithExtra adds an additional FFmpeg argument.
func (b *ProfileBuilder) WithExtra(key, value string) *ProfileBuilder {
	if b.profile.ExtraArgs == nil {
		b.profile.ExtraArgs = make(map[string]string)
	}
	b.profile.ExtraArgs[key] = value
	return b
}

// Build returns the completed Profile.
func (b *ProfileBuilder) Build() *Profile {
	return &b.profile
}
