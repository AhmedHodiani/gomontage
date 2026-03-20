package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

// MediaInfo holds metadata about a media file, extracted via ffprobe.
type MediaInfo struct {
	// Path is the file path that was probed.
	Path string

	// Duration is the total duration of the media file.
	Duration time.Duration

	// Format is the container format name (e.g. "mov,mp4,m4a,3gp,3g2,mj2").
	Format string

	// FormatLong is the long name of the format (e.g. "QuickTime / MOV").
	FormatLong string

	// Size is the file size in bytes.
	Size int64

	// BitRate is the overall bitrate in bits per second.
	BitRate int64

	// VideoStreams contains info about each video stream found.
	VideoStreams []VideoStreamInfo

	// AudioStreams contains info about each audio stream found.
	AudioStreams []AudioStreamInfo
}

// HasVideo returns true if the media file contains at least one video stream.
func (m *MediaInfo) HasVideo() bool {
	return len(m.VideoStreams) > 0
}

// HasAudio returns true if the media file contains at least one audio stream.
func (m *MediaInfo) HasAudio() bool {
	return len(m.AudioStreams) > 0
}

// VideoStreamInfo holds metadata about a single video stream.
type VideoStreamInfo struct {
	// Index is the stream index within the file.
	Index int

	// Codec is the codec name (e.g. "h264", "vp9", "av1").
	Codec string

	// CodecLong is the long codec name.
	CodecLong string

	// Width is the video width in pixels.
	Width int

	// Height is the video height in pixels.
	Height int

	// FPS is the frame rate as frames per second.
	FPS float64

	// Duration is the stream duration. May differ from container duration.
	Duration time.Duration

	// BitRate is the video bitrate in bits per second.
	BitRate int64

	// PixelFormat is the pixel format (e.g. "yuv420p").
	PixelFormat string
}

// AudioStreamInfo holds metadata about a single audio stream.
type AudioStreamInfo struct {
	// Index is the stream index within the file.
	Index int

	// Codec is the codec name (e.g. "aac", "mp3", "pcm_s16le").
	Codec string

	// CodecLong is the long codec name.
	CodecLong string

	// SampleRate is the audio sample rate in Hz.
	SampleRate int

	// Channels is the number of audio channels.
	Channels int

	// ChannelLayout is the channel layout (e.g. "stereo", "5.1").
	ChannelLayout string

	// Duration is the stream duration.
	Duration time.Duration

	// BitRate is the audio bitrate in bits per second.
	BitRate int64
}

// ffprobeOutput is the raw JSON structure returned by ffprobe.
type ffprobeOutput struct {
	Format  ffprobeFormat   `json:"format"`
	Streams []ffprobeStream `json:"streams"`
}

type ffprobeFormat struct {
	Filename       string `json:"filename"`
	FormatName     string `json:"format_name"`
	FormatLongName string `json:"format_long_name"`
	Duration       string `json:"duration"`
	Size           string `json:"size"`
	BitRate        string `json:"bit_rate"`
}

type ffprobeStream struct {
	Index         int    `json:"index"`
	CodecType     string `json:"codec_type"`
	CodecName     string `json:"codec_name"`
	CodecLongName string `json:"codec_long_name"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	RFrameRate    string `json:"r_frame_rate"`
	Duration      string `json:"duration"`
	BitRate       string `json:"bit_rate"`
	SampleRate    string `json:"sample_rate"`
	Channels      int    `json:"channels"`
	ChannelLayout string `json:"channel_layout"`
	PixelFormat   string `json:"pix_fmt"`
}

// Probe uses ffprobe to extract metadata from a media file.
// It returns a MediaInfo struct with all discovered streams and format info.
func Probe(path string) (*MediaInfo, error) {
	return ProbeContext(context.Background(), path)
}

// ProbeContext is like Probe but accepts a context for cancellation.
func ProbeContext(ctx context.Context, path string) (*MediaInfo, error) {
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		path,
	}

	cmd := exec.CommandContext(ctx, "ffprobe", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed for %q: %w", path, err)
	}

	var raw ffprobeOutput
	if err := json.Unmarshal(output, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output for %q: %w", path, err)
	}

	info := &MediaInfo{
		Path:       path,
		Format:     raw.Format.FormatName,
		FormatLong: raw.Format.FormatLongName,
	}

	// Parse format-level fields.
	if d, err := parseSeconds(raw.Format.Duration); err == nil {
		info.Duration = d
	}
	if v, err := strconv.ParseInt(raw.Format.Size, 10, 64); err == nil {
		info.Size = v
	}
	if v, err := strconv.ParseInt(raw.Format.BitRate, 10, 64); err == nil {
		info.BitRate = v
	}

	// Parse streams.
	for _, s := range raw.Streams {
		switch s.CodecType {
		case "video":
			vs := VideoStreamInfo{
				Index:       s.Index,
				Codec:       s.CodecName,
				CodecLong:   s.CodecLongName,
				Width:       s.Width,
				Height:      s.Height,
				PixelFormat: s.PixelFormat,
			}
			if d, err := parseSeconds(s.Duration); err == nil {
				vs.Duration = d
			}
			if v, err := strconv.ParseInt(s.BitRate, 10, 64); err == nil {
				vs.BitRate = v
			}
			vs.FPS = parseFraction(s.RFrameRate)
			info.VideoStreams = append(info.VideoStreams, vs)

		case "audio":
			as := AudioStreamInfo{
				Index:         s.Index,
				Codec:         s.CodecName,
				CodecLong:     s.CodecLongName,
				Channels:      s.Channels,
				ChannelLayout: s.ChannelLayout,
			}
			if d, err := parseSeconds(s.Duration); err == nil {
				as.Duration = d
			}
			if v, err := strconv.ParseInt(s.BitRate, 10, 64); err == nil {
				as.BitRate = v
			}
			if v, err := strconv.Atoi(s.SampleRate); err == nil {
				as.SampleRate = v
			}
			info.AudioStreams = append(info.AudioStreams, as)
		}
	}

	return info, nil
}

// parseSeconds parses a decimal seconds string (e.g. "123.456") into a Duration.
func parseSeconds(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("empty duration string")
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return time.Duration(f * float64(time.Second)), nil
}

// parseFraction parses a fraction string like "30000/1001" into a float64.
// Returns 0 if parsing fails.
func parseFraction(s string) float64 {
	if s == "" {
		return 0
	}
	// Try fraction format first (e.g. "30000/1001").
	var num, den float64
	if n, _ := fmt.Sscanf(s, "%f/%f", &num, &den); n == 2 && den != 0 {
		return num / den
	}
	// Fall back to plain float.
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return 0
}
