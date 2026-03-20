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
}

// ProgressFunc is a callback invoked with progress updates during encoding.
type ProgressFunc func(Progress)

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
	done := make(chan error, 1)
	go func() {
		parseProgress(stderr, onProgress)
		done <- proc.Wait()
	}()

	select {
	case <-ctx.Done():
		// Context cancelled — kill the process.
		_ = proc.Process.Kill()
		<-done // Wait for goroutine to finish.
		return ctx.Err()
	case err := <-done:
		if err != nil {
			return fmt.Errorf("ffmpeg failed: %w", err)
		}
		return nil
	}
}

// progressRegex matches FFmpeg progress output lines like:
// frame=  100 fps= 30 q=28.0 size=    1024kB time=00:00:03.33 bitrate=2515.2kbits/s speed=1.00x
var progressRegex = regexp.MustCompile(
	`frame=\s*(\d+)\s+fps=\s*([\d.]+)\s+.*size=\s*(\S+)\s+time=\s*(\S+)\s+bitrate=\s*(\S+)\s+speed=\s*(\S+)`,
)

// parseProgress reads FFmpeg stderr output and invokes the callback on each progress line.
func parseProgress(r io.Reader, onProgress ProgressFunc) {
	if onProgress == nil {
		// Drain the reader even if we don't need progress.
		_, _ = io.Copy(io.Discard, r)
		return
	}

	scanner := bufio.NewScanner(r)
	// FFmpeg uses \r for progress updates, so we need a custom split function.
	scanner.Split(splitOnCRorLF)

	for scanner.Scan() {
		line := scanner.Text()
		matches := progressRegex.FindStringSubmatch(line)
		if matches == nil {
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

		onProgress(p)
	}
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
