package timeline

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ahmedhodiani/gomontage/engine"
	"github.com/ahmedhodiani/gomontage/export"
)

// Export compiles the timeline into an FFmpeg command and renders the output
// to the specified file path using the given export profile.
//
// Export prints compilation status, a progress bar, and a summary to stderr.
// For silent operation (no output), use ExportSilent.
//
// This is the final step of a Gomontage workflow. Everything is lazy until
// this method is called — no FFmpeg processes run until Export.
//
// Example:
//
//	err := tl.Export(export.YouTube1080p(), "output/final.mp4")
func (tl *Timeline) Export(profile *export.Profile, outputPath string) error {
	return tl.ExportContext(context.Background(), profile, outputPath, nil)
}

// ExportSilent compiles and exports the timeline with no output.
// This is equivalent to the old Export behavior — completely silent.
//
// Example:
//
//	err := tl.ExportSilent(export.YouTube1080p(), "output/final.mp4")
func (tl *Timeline) ExportSilent(profile *export.Profile, outputPath string) error {
	return tl.exportInternal(context.Background(), profile, outputPath, nil, false)
}

// ExportWithProgress is like Export but reports progress via a callback
// instead of using the built-in progress bar. Compilation and summary
// lines are still printed to stderr.
//
// Example:
//
//	err := tl.ExportWithProgress(export.YouTube1080p(), "output/final.mp4",
//	    func(p engine.Progress) {
//	        fmt.Printf("frame=%d time=%s speed=%.1fx\n", p.Frame, p.Time, p.Speed)
//	    })
func (tl *Timeline) ExportWithProgress(profile *export.Profile, outputPath string, onProgress engine.ProgressFunc) error {
	return tl.ExportContext(context.Background(), profile, outputPath, onProgress)
}

// ExportContext is like Export but accepts a context for cancellation and a
// progress callback. This is the most flexible export method.
//
// If onProgress is nil, a built-in progress bar is displayed on stderr.
// If onProgress is non-nil, the callback receives progress updates instead.
func (tl *Timeline) ExportContext(ctx context.Context, profile *export.Profile, outputPath string, onProgress engine.ProgressFunc) error {
	return tl.exportInternal(ctx, profile, outputPath, onProgress, true)
}

// exportInternal is the shared implementation for all export methods.
// When verbose is true, compilation status, progress bar, and summary are
// printed to stderr. When verbose is false, the export runs silently.
func (tl *Timeline) exportInternal(ctx context.Context, profile *export.Profile, outputPath string, onProgress engine.ProgressFunc, verbose bool) error {
	exportStart := time.Now()

	// Step 1: Compile timeline into filter graph.
	if verbose {
		printTimelineSummary(tl)
	}

	compileStart := time.Now()
	compiler := NewCompiler(tl)
	graph, err := compiler.Compile(outputPath, profile.Params())
	if err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}

	if verbose {
		compileDur := time.Since(compileStart)
		fmt.Fprintf(os.Stderr, "gomontage: Timeline compiled (%s)\n", formatCompileDuration(compileDur))
	}

	// Step 2: Build FFmpeg command from the graph.
	cmd, err := engine.BuildCommand(graph)
	if err != nil {
		return fmt.Errorf("command build failed: %w", err)
	}

	if verbose {
		printCommandSummary(graph, outputPath)
	}

	// Step 3: Execute the FFmpeg command.
	totalDuration := tl.Duration()
	opts := engine.RunOptions{
		TotalDuration: totalDuration,
	}

	if onProgress != nil {
		// User provided their own callback — use it directly.
		opts.OnProgress = onProgress
	} else if verbose {
		// No user callback and verbose mode — show progress bar.
		bar := engine.NewProgressBar(totalDuration, os.Stderr)
		opts.OnProgress = func(p engine.Progress) {
			bar.Update(p)
		}
		defer func() {
			if err == nil {
				// Get output file size.
				var fileSize int64
				if info, statErr := os.Stat(outputPath); statErr == nil {
					fileSize = info.Size()
				}
				elapsed := time.Since(exportStart)
				bar.Finish(outputPath, fileSize, elapsed)
			} else {
				// Clear the progress bar line on error.
				fmt.Fprintf(os.Stderr, "\r\033[K")
			}
		}()
	}

	if err := engine.RunContextOpts(ctx, cmd, opts); err != nil {
		return fmt.Errorf("export failed: %w", err)
	}

	return nil
}

// DryRun compiles the timeline and returns the FFmpeg command that would be
// executed, without actually running it. Useful for debugging and previewing
// the generated command.
//
// Example:
//
//	cmd, err := tl.DryRun(export.YouTube1080p(), "output/final.mp4")
//	fmt.Println(cmd.String())  // Prints: ffmpeg -y -i input.mp4 ...
func (tl *Timeline) DryRun(profile *export.Profile, outputPath string) (*engine.Command, error) {
	compiler := NewCompiler(tl)
	graph, err := compiler.Compile(outputPath, profile.Params())
	if err != nil {
		return nil, fmt.Errorf("compilation failed: %w", err)
	}

	cmd, err := engine.BuildCommand(graph)
	if err != nil {
		return nil, fmt.Errorf("command build failed: %w", err)
	}

	return cmd, nil
}

// ExportRange exports only the portion of the timeline between start and end.
// Clips outside the range are excluded, partially overlapping clips are trimmed,
// and the output starts at t=0. This is useful for quickly previewing a section
// of a complex timeline without rendering the entire project.
//
// Example:
//
//	// Export only minutes 5-7 of a long timeline.
//	err := tl.ExportRange(5*time.Minute, 7*time.Minute, export.YouTube1080p(), "preview.mp4")
func (tl *Timeline) ExportRange(start, end time.Duration, profile *export.Profile, outputPath string) error {
	sub := tl.SubRange(start, end)
	return sub.Export(profile, outputPath)
}

// ExportRangeSilent is like ExportRange but produces no output.
func (tl *Timeline) ExportRangeSilent(start, end time.Duration, profile *export.Profile, outputPath string) error {
	sub := tl.SubRange(start, end)
	return sub.ExportSilent(profile, outputPath)
}

// DryRunRange is like DryRun but only includes the portion of the timeline
// between start and end.
func (tl *Timeline) DryRunRange(start, end time.Duration, profile *export.Profile, outputPath string) (*engine.Command, error) {
	sub := tl.SubRange(start, end)
	return sub.DryRun(profile, outputPath)
}

// printTimelineSummary prints a one-line summary of the timeline to stderr.
func printTimelineSummary(tl *Timeline) {
	cfg := tl.Config()
	vTracks := len(tl.VideoTracks())
	aTracks := len(tl.AudioTracks())

	totalClips := 0
	for _, t := range tl.VideoTracks() {
		totalClips += len(t.Entries())
	}
	for _, t := range tl.AudioTracks() {
		totalClips += len(t.Entries())
	}

	dur := tl.Duration()

	fmt.Fprintf(os.Stderr, "gomontage: %dx%d @ %.0ffps | %s, %s | %d clips | %s total\n",
		cfg.Width, cfg.Height, cfg.FPS,
		pluralize(vTracks, "video track"),
		pluralize(aTracks, "audio track"),
		totalClips,
		engine.FormatDurationShort(dur),
	)
}

// printCommandSummary prints a one-line FFmpeg command summary to stderr.
func printCommandSummary(graph *engine.Graph, outputPath string) {
	inputCount := len(graph.Inputs())

	filterCount := 0
	for _, node := range graph.Nodes() {
		if node.Type == engine.NodeFilter {
			filterCount++
		}
	}

	fmt.Fprintf(os.Stderr, "gomontage: FFmpeg command: %s, %s -> %s\n",
		pluralize(inputCount, "input"),
		pluralize(filterCount, "filter"),
		outputPath,
	)
}

// formatCompileDuration formats a compile duration for display.
func formatCompileDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	return fmt.Sprintf("%dms", d.Milliseconds())
}

// pluralize returns "N thing" or "N things" based on count.
func pluralize(n int, singular string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, singular)
	}
	return fmt.Sprintf("%d %ss", n, singular)
}
