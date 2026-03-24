package engine

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Progress reports the current state of an FFmpeg encoding process.
type Progress struct {
	// Frame is the current frame number being processed.
	Frame int

	// FPS is the current encoding speed in frames per second.
	FPS float64

	// Size is the current output file size as a human-readable string.
	Size string

	// Time is how far into the output we've encoded.
	Time time.Duration

	// Bitrate is the current encoding bitrate as a string (e.g. "5000kbits/s").
	Bitrate string

	// Speed is the encoding speed relative to real-time (e.g. 2.5 means 2.5x faster).
	Speed float64

	// Percent is the estimated completion percentage (0-100).
	// Only set when TotalDuration is provided via RunOptions.
	Percent float64

	// ETA is the estimated time remaining until the export completes.
	// Only set when TotalDuration is provided and Speed > 0.
	ETA time.Duration
}

// ProgressFunc is a callback invoked with progress updates during encoding.
type ProgressFunc func(Progress)

// RunOptions configures how an FFmpeg command is executed.
type RunOptions struct {
	// OnProgress is a callback invoked with progress updates during encoding.
	// If nil, progress is not reported.
	OnProgress ProgressFunc

	// TotalDuration is the expected total duration of the output.
	// When set, Progress.Percent and Progress.ETA are computed automatically.
	TotalDuration time.Duration
}

// Run executes an FFmpeg command and blocks until it completes.
// Returns an error if FFmpeg exits with a non-zero status.
func Run(cmd *Command) error {
	return RunContext(context.Background(), cmd, nil)
}

// RunWithProgress executes an FFmpeg command and reports progress via the callback.
func RunWithProgress(cmd *Command, onProgress ProgressFunc) error {
	return RunContext(context.Background(), cmd, onProgress)
}

// RunContext executes an FFmpeg command with context support and optional progress reporting.
func RunContext(ctx context.Context, cmd *Command, onProgress ProgressFunc) error {
	return RunContextOpts(ctx, cmd, RunOptions{OnProgress: onProgress})
}

// RunContextOpts executes an FFmpeg command with full control over progress
// reporting, ETA calculation, and stderr capture.
func RunContextOpts(ctx context.Context, cmd *Command, opts RunOptions) error {
	proc := exec.CommandContext(ctx, cmd.Binary, cmd.Args...)

	// FFmpeg writes progress info to stderr.
	stderr, err := proc.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to capture stderr: %w", err)
	}

	if err := proc.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Parse progress from stderr in a goroutine.
	type result struct {
		err        error
		stderrTail []string
	}
	done := make(chan result, 1)
	go func() {
		tail := parseProgressOpts(stderr, opts)
		waitErr := proc.Wait()
		done <- result{err: waitErr, stderrTail: tail}
	}()

	select {
	case <-ctx.Done():
		// Context cancelled — kill the process.
		_ = proc.Process.Kill()
		<-done // Wait for goroutine to finish.
		return ctx.Err()
	case r := <-done:
		if r.err != nil {
			return newFFmpegError(r.err, r.stderrTail)
		}
		return nil
	}
}

// maxStderrLines is the maximum number of non-progress stderr lines to keep
// in the ring buffer for error reporting.
const maxStderrLines = 25

// FFmpegError wraps an FFmpeg process error with captured stderr output
// for debugging.
type FFmpegError struct {
	Err        error
	StderrTail []string
}

func (e *FFmpegError) Error() string {
	if len(e.StderrTail) == 0 {
		return fmt.Sprintf("ffmpeg failed: %v", e.Err)
	}
	return fmt.Sprintf("ffmpeg failed: %v\n%s", e.Err, strings.Join(e.StderrTail, "\n"))
}

func (e *FFmpegError) Unwrap() error {
	return e.Err
}

func newFFmpegError(err error, stderrTail []string) *FFmpegError {
	return &FFmpegError{Err: err, StderrTail: stderrTail}
}

// progressRegex matches FFmpeg progress output lines like:
// frame=  100 fps= 30 q=28.0 size=    1024kB time=00:00:03.33 bitrate=2515.2kbits/s speed=1.00x
var progressRegex = regexp.MustCompile(
	`frame=\s*(\d+)\s+fps=\s*([\d.]+)\s+.*size=\s*(\S+)\s+time=\s*(\S+)\s+bitrate=\s*(\S+)\s+speed=\s*(\S+)`,
)

// parseProgressOpts reads FFmpeg stderr output, invokes the progress callback,
// computes Percent/ETA when TotalDuration is set, and returns the last N
// non-progress lines for error reporting.
func parseProgressOpts(r io.Reader, opts RunOptions) []string {
	// Ring buffer for non-progress stderr lines.
	stderrBuf := make([]string, 0, maxStderrLines)

	if opts.OnProgress == nil {
		// Even without a progress callback, capture stderr for error reporting.
		scanner := bufio.NewScanner(r)
		scanner.Split(splitOnCRorLF)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			// Skip progress lines from the buffer — they're not useful for errors.
			if progressRegex.MatchString(line) {
				continue
			}
			stderrBuf = appendRing(stderrBuf, line, maxStderrLines)
		}
		return stderrBuf
	}

	scanner := bufio.NewScanner(r)
	// FFmpeg uses \r for progress updates, so we need a custom split function.
	scanner.Split(splitOnCRorLF)

	for scanner.Scan() {
		line := scanner.Text()
		matches := progressRegex.FindStringSubmatch(line)
		if matches == nil {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				stderrBuf = appendRing(stderrBuf, trimmed, maxStderrLines)
			}
			continue
		}

		p := Progress{
			Size:    matches[3],
			Bitrate: matches[5],
		}

		if v, err := strconv.Atoi(matches[1]); err == nil {
			p.Frame = v
		}
		if v, err := strconv.ParseFloat(matches[2], 64); err == nil {
			p.FPS = v
		}
		if d, err := parseFFmpegTime(matches[4]); err == nil {
			p.Time = d
		}
		speedStr := strings.TrimSuffix(matches[6], "x")
		if v, err := strconv.ParseFloat(speedStr, 64); err == nil {
			p.Speed = v
		}

		// Compute Percent and ETA when total duration is known.
		if opts.TotalDuration > 0 && p.Time > 0 {
			p.Percent = float64(p.Time) / float64(opts.TotalDuration) * 100
			if p.Percent > 100 {
				p.Percent = 100
			}
			if p.Speed > 0 {
				remaining := opts.TotalDuration - p.Time
				if remaining > 0 {
					p.ETA = time.Duration(float64(remaining) / p.Speed)
				}
			}
		}

		opts.OnProgress(p)
	}

	return stderrBuf
}

// appendRing appends a line to a ring buffer, dropping the oldest entry
// when the buffer is full.
func appendRing(buf []string, line string, max int) []string {
	if len(buf) < max {
		return append(buf, line)
	}
	// Shift left and replace last.
	copy(buf, buf[1:])
	buf[max-1] = line
	return buf
}

// parseFFmpegTime parses an FFmpeg time string like "00:01:23.45" into a Duration.
func parseFFmpegTime(s string) (time.Duration, error) {
	// Handle negative times (shouldn't happen but be safe).
	s = strings.TrimPrefix(s, "-")

	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid time format: %s", s)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	seconds, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return 0, err
	}

	d := time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(seconds*float64(time.Second))

	return d, nil
}

// splitOnCRorLF is a bufio.SplitFunc that splits on \r or \n,
// which is needed because FFmpeg uses \r for in-place progress updates.
func splitOnCRorLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	for i, b := range data {
		if b == '\n' || b == '\r' {
			return i + 1, data[:i], nil
		}
	}

	if atEOF {
		return len(data), data, nil
	}

	// Request more data.
	return 0, nil, nil
}
