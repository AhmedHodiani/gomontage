package main

import (
	"fmt"
	"time"

	"github.com/ahmedhodiani/gomontage/engine"
	"github.com/spf13/cobra"
)

func probeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "probe <file>",
		Short: "Inspect a media file's properties",
		Long: `Uses ffprobe to extract and display metadata about a media file,
including format, duration, resolution, codecs, bitrate, and stream info.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			info, err := engine.Probe(path)
			if err != nil {
				return fmt.Errorf("probe failed: %w", err)
			}

			fmt.Printf("File:     %s\n", info.Path)
			fmt.Printf("Format:   %s (%s)\n", info.Format, info.FormatLong)
			fmt.Printf("Duration: %s\n", formatDuration(info.Duration))
			fmt.Printf("Size:     %s\n", formatBytes(info.Size))

			if info.BitRate > 0 {
				fmt.Printf("Bitrate:  %s\n", formatBitrate(info.BitRate))
			}

			for i, vs := range info.VideoStreams {
				fmt.Printf("\nVideo Stream #%d:\n", i)
				fmt.Printf("  Codec:        %s (%s)\n", vs.Codec, vs.CodecLong)
				fmt.Printf("  Resolution:   %dx%d\n", vs.Width, vs.Height)
				fmt.Printf("  Frame Rate:   %.2f fps\n", vs.FPS)
				fmt.Printf("  Pixel Format: %s\n", vs.PixelFormat)
				if vs.BitRate > 0 {
					fmt.Printf("  Bitrate:      %s\n", formatBitrate(vs.BitRate))
				}
			}

			for i, as := range info.AudioStreams {
				fmt.Printf("\nAudio Stream #%d:\n", i)
				fmt.Printf("  Codec:       %s (%s)\n", as.Codec, as.CodecLong)
				fmt.Printf("  Sample Rate: %d Hz\n", as.SampleRate)
				fmt.Printf("  Channels:    %d", as.Channels)
				if as.ChannelLayout != "" {
					fmt.Printf(" (%s)", as.ChannelLayout)
				}
				fmt.Println()
				if as.BitRate > 0 {
					fmt.Printf("  Bitrate:     %s\n", formatBitrate(as.BitRate))
				}
			}

			return nil
		},
	}
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := d.Seconds() - float64(h*3600+m*60)

	if h > 0 {
		return fmt.Sprintf("%dh %dm %.1fs", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %.1fs", m, s)
	}
	return fmt.Sprintf("%.1fs", s)
}

func formatBytes(b int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.2f GB", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.2f MB", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.2f KB", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func formatBitrate(bps int64) string {
	if bps >= 1_000_000 {
		return fmt.Sprintf("%.1f Mbps", float64(bps)/1_000_000)
	}
	return fmt.Sprintf("%.0f kbps", float64(bps)/1_000)
}
