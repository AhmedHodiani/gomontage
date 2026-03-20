package timeline

import (
	"context"
	"fmt"

	"github.com/ahmedhodiani/gomontage/engine"
	"github.com/ahmedhodiani/gomontage/export"
)

// Export compiles the timeline into an FFmpeg command and renders the output
// to the specified file path using the given export profile.
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

// ExportWithProgress is like Export but reports progress via a callback.
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
func (tl *Timeline) ExportContext(ctx context.Context, profile *export.Profile, outputPath string, onProgress engine.ProgressFunc) error {
	// Step 1: Compile timeline into filter graph.
	compiler := NewCompiler(tl)
	graph, err := compiler.Compile(outputPath, profile.Params())
	if err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}

	// Step 2: Build FFmpeg command from the graph.
	cmd, err := engine.BuildCommand(graph)
	if err != nil {
		return fmt.Errorf("command build failed: %w", err)
	}

	// Step 3: Execute the FFmpeg command.
	if err := engine.RunContext(ctx, cmd, onProgress); err != nil {
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
