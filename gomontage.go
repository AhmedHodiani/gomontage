// Package gomontage is a programmatic video editing framework for Go.
//
// Gomontage lets you describe video edits with high-level Go code using a
// track-based timeline model. Load clips, arrange them on tracks, add
// transitions and effects, layer audio, and export — all in clean,
// readable Go.
//
// # Quick Start
//
//	tl := gomontage.NewTimeline(gomontage.TimelineConfig{
//	    Width:  1920,
//	    Height: 1080,
//	    FPS:    30,
//	})
//
//	video := tl.AddVideoTrack("main")
//	video.Add(clip.NewVideo("intro.mp4"), gomontage.At(0))
//
//	tl.Export(export.YouTube1080p(), "output/final.mp4")
//
// See the subpackages for detailed documentation:
//   - clip: Media clip types (video, audio, image, text, color)
//   - timeline: Track-based timeline and composition
//   - cuts: Transition types (L-cut, J-cut, dissolve, etc.)
//   - effects: Audio and video effects (fade, volume, speed)
//   - export: Output profiles and presets
//   - engine: Low-level FFmpeg interface (internal)
package gomontage
