package engine

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

// ProgressBar renders a single-line progress bar to a writer.
//
// Example output:
//
//	Exporting  ████████████░░░░░░░░  58% | 05:48 / 10:00 | ETA 4m12s | 2.3x
type ProgressBar struct {
	total   time.Duration
	writer  io.Writer
	start   time.Time
	lastLen int // length of last rendered line, for clearing
}

// NewProgressBar creates a progress bar that renders to w.
// If total is zero, the bar shows elapsed time and speed without percentage.
func NewProgressBar(total time.Duration, w io.Writer) *ProgressBar {
	if w == nil {
		w = os.Stderr
	}
	return &ProgressBar{
		total:  total,
		writer: w,
		start:  time.Now(),
	}
}

// Update redraws the progress bar with the latest progress data.
func (pb *ProgressBar) Update(p Progress) {
	width := pb.termWidth()

	var line string
	if pb.total > 0 && p.Percent > 0 {
		line = pb.renderWithPercent(p, width)
	} else {
		line = pb.renderElapsedOnly(p, width)
	}

	// Pad with spaces to clear any leftover characters from the previous line.
	if len(line) < pb.lastLen {
		line += strings.Repeat(" ", pb.lastLen-len(line))
	}
	pb.lastLen = len(line)

	fmt.Fprintf(pb.writer, "\r%s", line)
}

// Finish prints the final export summary and moves to a new line.
func (pb *ProgressBar) Finish(outputPath string, fileSize int64, elapsed time.Duration) {
	// Clear the progress line.
	if pb.lastLen > 0 {
		fmt.Fprintf(pb.writer, "\r%s\r", strings.Repeat(" ", pb.lastLen))
	}

	sizeStr := formatFileSize(fileSize)
	fmt.Fprintf(pb.writer, "gomontage: Export complete: %s (%s) in %s\n",
		outputPath, sizeStr, formatDurationHuman(elapsed))
}

// renderWithPercent renders the bar when we know the total duration:
//
//	Exporting  ████████░░░░░░░░  45% | 04:30 / 10:00 | ETA 5m30s | 2.1x
func (pb *ProgressBar) renderWithPercent(p Progress, width int) string {
	prefix := "Exporting  "
	pct := fmt.Sprintf(" %3.0f%%", p.Percent)
	timeInfo := fmt.Sprintf(" | %s / %s", FormatDurationShort(p.Time), FormatDurationShort(pb.total))

	var etaStr string
	if p.ETA > 0 {
		etaStr = fmt.Sprintf(" | ETA %s", formatDurationHuman(p.ETA))
	}

	var speedStr string
	if p.Speed > 0 {
		speedStr = fmt.Sprintf(" | %.1fx", p.Speed)
	}

	suffix := pct + timeInfo + etaStr + speedStr

	// Calculate bar width from remaining space.
	barWidth := width - len(prefix) - len(suffix)
	if barWidth < 5 {
		// Terminal too narrow for a bar — just show text.
		return prefix + suffix
	}

	bar := renderBar(p.Percent, barWidth)
	return prefix + bar + suffix
}

// renderElapsedOnly renders the bar when total duration is unknown:
//
//	Exporting  ░░░░░░░░░░░░░░░░ | 02:15 elapsed | 1.8x
func (pb *ProgressBar) renderElapsedOnly(p Progress, width int) string {
	elapsed := time.Since(pb.start).Truncate(time.Second)
	parts := []string{
		"Exporting",
		fmt.Sprintf(" | %s elapsed", FormatDurationShort(elapsed)),
	}
	if p.Time > 0 {
		parts = append(parts, fmt.Sprintf(" | at %s", FormatDurationShort(p.Time)))
	}
	if p.Speed > 0 {
		parts = append(parts, fmt.Sprintf(" | %.1fx", p.Speed))
	}
	if p.Size != "" && p.Size != "N/A" {
		parts = append(parts, fmt.Sprintf(" | %s", p.Size))
	}
	line := strings.Join(parts, "")
	if len(line) > width {
		line = line[:width]
	}
	return line
}

// renderBar builds a progress bar string of the given width using block characters.
func renderBar(percent float64, width int) string {
	if width <= 0 {
		return ""
	}
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

// termWidth returns the terminal width, defaulting to 80 if it can't be determined.
func (pb *ProgressBar) termWidth() int {
	if f, ok := pb.writer.(*os.File); ok {
		if w, _, err := term.GetSize(int(f.Fd())); err == nil && w > 0 {
			return w
		}
	}
	return 80
}

// FormatDurationShort formats a duration as MM:SS or HH:MM:SS.
func FormatDurationShort(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	d = d.Truncate(time.Second)

	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

// formatDurationHuman formats a duration as a human-friendly string like "1m23s" or "45s".
func formatDurationHuman(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	d = d.Truncate(time.Second)

	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		m := int(d.Minutes())
		s := int(d.Seconds()) % 60
		if s > 0 {
			return fmt.Sprintf("%dm%ds", m, s)
		}
		return fmt.Sprintf("%dm", m)
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if m > 0 {
		return fmt.Sprintf("%dh%dm", h, m)
	}
	return fmt.Sprintf("%dh", h)
}

// formatFileSize formats a byte count as a human-readable string.
func formatFileSize(bytes int64) string {
	if bytes < 0 {
		return "0 B"
	}

	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
	)

	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
