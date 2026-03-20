// Package docs provides automated API documentation generation for Gomontage.
//
// The generator produces markdown files covering all public APIs of the
// framework. These docs are the primary reference for users writing
// Gomontage code.
//
// Usage:
//
//	gomontage docs              # Generates docs/ directory
//	gomontage docs -o api-docs  # Custom output directory
package docs

import (
	"fmt"
	"os"
	"path/filepath"
)

// Generate produces markdown documentation files in the specified output directory.
func Generate(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("could not create docs directory: %w", err)
	}

	generators := []struct {
		filename string
		content  string
	}{
		{"getting-started.md", gettingStartedDoc()},
		{"clips.md", clipsDoc()},
		{"timeline.md", timelineDoc()},
		{"cuts.md", cutsDoc()},
		{"effects.md", effectsDoc()},
		{"export.md", exportDoc()},
		{"api-reference.md", apiReferenceDoc()},
	}

	for _, gen := range generators {
		path := filepath.Join(outputDir, gen.filename)
		if err := os.WriteFile(path, []byte(gen.content), 0644); err != nil {
			return fmt.Errorf("could not write %s: %w", gen.filename, err)
		}
		fmt.Printf("  Generated %s\n", gen.filename)
	}

	return nil
}

func gettingStartedDoc() string {
	return `# Getting Started with Gomontage

Gomontage is a programmatic video editing framework for Go. Write Go code
to edit videos instead of using a GUI.

## Installation

` + "```" + `bash
go install github.com/ahmedhodiani/gomontage/cmd/gomontage@latest
` + "```" + `

## Create a New Project

` + "```" + `bash
gomontage init my-video
cd my-video
` + "```" + `

This creates the following structure:

` + "```" + `
my-video/
├── gomontage.yaml       # Project configuration
├── main.go              # Your editing script
├── go.mod               # Go module
├── resources/
│   ├── video/           # Place video files here
│   ├── audio/           # Place audio files here
│   ├── images/          # Place images here
│   └── fonts/           # Place fonts here
├── output/              # Rendered output
└── temp/                # Temporary files
` + "```" + `

## Your First Edit

Edit ` + "`main.go`" + ` to describe your video:

` + "```go" + `
package main

import (
    "time"
    "github.com/ahmedhodiani/gomontage/clip"
    "github.com/ahmedhodiani/gomontage/timeline"
    "github.com/ahmedhodiani/gomontage/export"
)

func main() {
    tl := timeline.New(timeline.Config{
        Width:  1920,
        Height: 1080,
        FPS:    30,
    })

    video := clip.NewVideo("resources/video/footage.mp4")
    trimmed := video.Trim(0, 30*time.Second)

    main := tl.AddVideoTrack("main")
    main.Add(trimmed, timeline.At(0))

    tl.Export(export.YouTube1080p(), "output/final.mp4")
}
` + "```" + `

## Run Your Project

` + "```" + `bash
gomontage run
` + "```" + `

## Inspect Media Files

` + "```" + `bash
gomontage probe resources/video/footage.mp4
` + "```" + `

## Key Concepts

1. **Clips** are your media building blocks (video, audio, image, text, color)
2. **Tracks** are named layers on the timeline (video tracks, audio tracks)
3. **Timeline** is the master container that holds all tracks
4. **Cuts** define transitions between adjacent clips
5. **Effects** are transformations applied to clips
6. **Export profiles** configure the output format and quality
`
}

func clipsDoc() string {
	return `# Clips

Clips are the fundamental building blocks of a Gomontage project. Every clip
is **immutable** — transform methods return new clips, leaving originals unchanged.

## Clip Types

### VideoClip

Load video from a file. Contains both video and audio streams.

` + "```go" + `
video := clip.NewVideo("resources/video/interview.mp4")

// Trim to a segment (times relative to source)
segment := video.Trim(10*time.Second, 30*time.Second)

// Adjust audio volume
quiet := video.WithVolume(0.5)

// Add fades
faded := video.WithFadeIn(1*time.Second).WithFadeOut(2*time.Second)

// Use only video (remove audio)
videoOnly := video.VideoOnly()

// Use only audio (remove video)
audioOnly := video.AudioOnly()

// Resize
small := video.WithSize(640, 360)
` + "```" + `

### AudioClip

Load audio from a file (WAV, MP3, FLAC, etc.).

` + "```go" + `
narration := clip.NewAudio("resources/audio/narration.wav")
music := clip.NewAudio("resources/audio/bgm.mp3").WithVolume(0.3)

// Trim
intro := narration.Trim(0, 30*time.Second)

// Fade in/out
faded := music.WithFadeIn(3*time.Second).WithFadeOut(5*time.Second)
` + "```" + `

### ImageClip

Static images used as video clips. Default duration is 5 seconds.

` + "```go" + `
logo := clip.NewImage("resources/images/logo.png").WithDuration(3*time.Second)
bg := clip.NewImage("resources/images/background.jpg").WithDuration(10*time.Second)
` + "```" + `

### TextClip

Dynamically rendered text overlays.

` + "```go" + `
title := clip.NewText("Chapter 1", clip.TextStyle{
    Font:  "resources/fonts/bold.ttf",
    Size:  72,
    Color: "#FFFFFF",
}).WithDuration(4*time.Second).WithFadeIn(500*time.Millisecond)
` + "```" + `

### ColorClip

Solid color rectangles for backgrounds or transitions.

` + "```go" + `
black := clip.NewColor("#000000", 1920, 1080).WithDuration(2*time.Second)
` + "```" + `

## Immutability

All transforms return new clips. The original is never modified:

` + "```go" + `
original := clip.NewVideo("video.mp4")
trimmed := original.Trim(0, 10*time.Second)
// original is still the full video
// trimmed is a new clip with just the first 10 seconds
` + "```" + `

## Chaining

Transforms can be chained fluently:

` + "```go" + `
result := clip.NewVideo("raw.mp4").
    Trim(5*time.Second, 25*time.Second).
    WithVolume(0.8).
    WithFadeIn(1*time.Second).
    WithFadeOut(2*time.Second)
` + "```" + `
`
}

func timelineDoc() string {
	return `# Timeline & Tracks

The Timeline is the master container for your video project. It holds named
tracks where clips are placed at specific time positions.

## Creating a Timeline

` + "```go" + `
tl := timeline.New(timeline.Config{
    Width:  1920,
    Height: 1080,
    FPS:    30,
})
` + "```" + `

## Tracks

### Video Tracks

Video tracks hold visual clips. Multiple tracks are composited (layered)
in order — later tracks appear on top.

` + "```go" + `
main := tl.AddVideoTrack("main")       // Base layer
overlay := tl.AddVideoTrack("overlay")  // On top
` + "```" + `

### Audio Tracks

Audio tracks hold audio clips. All audio tracks are mixed together.

` + "```go" + `
voice := tl.AddAudioTrack("narration")
music := tl.AddAudioTrack("music")
sfx := tl.AddAudioTrack("sfx")
` + "```" + `

## Placing Clips

### At a Specific Time

` + "```go" + `
track.Add(myClip, timeline.At(5*time.Second))
` + "```" + `

### Sequentially (Back-to-Back)

` + "```go" + `
track.AddSequence(intro, mainContent, outro)
// intro starts at 0, mainContent starts when intro ends, etc.
` + "```" + `

## Timeline Duration

Duration is automatically calculated from the latest-ending clip:

` + "```go" + `
duration := tl.Duration() // Returns the total timeline length
` + "```" + `

## Validation

Check for errors before exporting:

` + "```go" + `
if err := tl.Validate(); err != nil {
    log.Fatal(err)
}
` + "```" + `
`
}

func cutsDoc() string {
	return `# Cuts & Transitions

Cuts define how two clips transition from one to the next.

## Adding Transitions

` + "```go" + `
// Between specific clips
tl.AddTransition(cuts.Dissolve(1*time.Second), clip1, clip2)

// Same transition for all adjacent clips on a track
track.TransitionAll(cuts.JCut(500*time.Millisecond))
` + "```" + `

## Cut Types

### Hard Cut
Instant switch. The default when no transition is specified.
` + "```go" + `
cuts.Hard()
` + "```" + `

### L-Cut
Audio from the outgoing clip continues over the incoming clip's video.
The edit forms an "L" shape on the timeline.
` + "```go" + `
cuts.LCut(2*time.Second) // Audio extends 2s into next clip
` + "```" + `

### J-Cut
Audio from the incoming clip starts before its video appears.
The edit forms a "J" shape. Builds anticipation.
` + "```go" + `
cuts.JCut(1*time.Second) // Audio starts 1s early
` + "```" + `

### Dissolve
Crossfade between two video clips.
` + "```go" + `
cuts.Dissolve(1*time.Second)
` + "```" + `

### CrossFade
Crossfade both video AND audio simultaneously.
` + "```go" + `
cuts.CrossFade(1*time.Second)
` + "```" + `

### Jump Cut
Sharp, jarring cut. No blending.
` + "```go" + `
cuts.JumpCut()
` + "```" + `

### Dip to Black / White
Fade out to a color, then fade in from that color.
` + "```go" + `
cuts.DipToBlack(1*time.Second)
cuts.DipToWhite(500*time.Millisecond)
` + "```" + `

### Wipe
Directional reveal of the incoming clip.
` + "```go" + `
cuts.Wipe(cuts.WipeLeft, 1*time.Second)
cuts.Wipe(cuts.WipeRight, 1*time.Second)
cuts.Wipe(cuts.WipeUp, 1*time.Second)
cuts.Wipe(cuts.WipeDown, 1*time.Second)
` + "```" + `
`
}

func effectsDoc() string {
	return `# Effects

Composable effects that transform clips.

## Video Effects

### Fade In / Fade Out
` + "```go" + `
effects.FadeIn(1*time.Second)
effects.FadeOut(2*time.Second)
effects.FadeOutAt(10*time.Second, 2*time.Second) // Start at specific time
` + "```" + `

### Speed
` + "```go" + `
effects.SpeedUp(2.0)   // 2x speed (timelapse)
effects.SpeedUp(0.5)   // Half speed (slow motion)
effects.SlowDown(2.0)  // Same as SpeedUp(0.5)
` + "```" + `

## Audio Effects

### Audio Fade
` + "```go" + `
effects.AudioFadeIn(3*time.Second)
effects.AudioFadeOut(5*time.Second)
` + "```" + `

### Volume
` + "```go" + `
effects.Volume(0.5)  // Half volume
effects.Volume(0.0)  // Mute
effects.Volume(2.0)  // Double volume
` + "```" + `

### Normalize
` + "```go" + `
effects.Normalize()          // Target -16 LUFS (broadcast standard)
effects.NormalizeTo(-23.0)   // Custom LUFS target
` + "```" + `

### Audio Speed
` + "```go" + `
effects.AudioSpeed(1.5)  // 1.5x playback speed
` + "```" + `
`
}

func exportDoc() string {
	return `# Export Profiles

Export profiles configure the output format, codec, quality, and resolution.

## Built-in Presets

### YouTube
` + "```go" + `
export.YouTube1080p()  // 1920x1080, H.264, AAC, CRF 18
export.YouTube4K()     // 3840x2160, H.264, AAC, CRF 18
export.YouTube720p()   // 1280x720, H.264, AAC, CRF 20
` + "```" + `

### Social Media
` + "```go" + `
export.Reel()  // 1080x1920 (9:16 vertical), H.264
` + "```" + `

### Professional
` + "```go" + `
export.ProRes()  // Apple ProRes 422, MOV container
export.H265()    // H.265/HEVC, better compression
` + "```" + `

### Audio Only
` + "```go" + `
export.MP3()   // MP3, 192kbps
export.WAV()   // Uncompressed PCM
export.FLAC()  // Lossless compressed
` + "```" + `

### Animated
` + "```go" + `
export.GIF()  // Animated GIF
` + "```" + `

## Custom Profiles

Use the fluent builder for full control:

` + "```go" + `
profile := export.NewProfile().
    WithName("Cinema 4K").
    WithCodec("libx265").
    WithCRF(18).
    WithPreset("slow").
    WithPixelFormat("yuv420p").
    WithResolution(3840, 2160).
    WithMaxRate("20M").
    WithBufSize("40M").
    WithAudioCodec("aac").
    WithAudioBitrate("320k").
    WithAudioSampleRate(48000).
    WithFastStart().
    Build()
` + "```" + `
`
}

func apiReferenceDoc() string {
	return `# API Reference

This is a quick reference for all Gomontage packages and their main types.

## Packages

| Package | Description |
|---------|-------------|
| ` + "`clip`" + ` | Media clip types (Video, Audio, Image, Text, Color) |
| ` + "`timeline`" + ` | Track-based timeline and composition |
| ` + "`cuts`" + ` | Transition types between clips |
| ` + "`effects`" + ` | Audio and video effects |
| ` + "`export`" + ` | Output profiles and presets |
| ` + "`engine`" + ` | Low-level FFmpeg interface (internal) |

## clip

| Type | Description |
|------|-------------|
| ` + "`VideoClip`" + ` | Video file clip with video + audio streams |
| ` + "`AudioClip`" + ` | Audio-only clip (WAV, MP3, FLAC, etc.) |
| ` + "`ImageClip`" + ` | Static image as video clip |
| ` + "`TextClip`" + ` | Rendered text overlay |
| ` + "`ColorClip`" + ` | Solid color fill |

### Common Methods (all clips)
- ` + "`.Duration()`" + ` — clip length
- ` + "`.Trim(start, end)`" + ` — extract a segment
- ` + "`.WithVolume(level)`" + ` — adjust audio volume
- ` + "`.WithFadeIn(d)`" + ` — add fade-in
- ` + "`.WithFadeOut(d)`" + ` — add fade-out

## timeline

| Type | Description |
|------|-------------|
| ` + "`Timeline`" + ` | Master container |
| ` + "`VideoTrack`" + ` | Visual clip layer |
| ` + "`AudioTrack`" + ` | Audio clip layer |

### Key Methods
- ` + "`timeline.New(config)`" + ` — create timeline
- ` + "`.AddVideoTrack(name)`" + ` — add video track
- ` + "`.AddAudioTrack(name)`" + ` — add audio track
- ` + "`track.Add(clip, At(time))`" + ` — place clip at time
- ` + "`track.AddSequence(clips...)`" + ` — place clips back-to-back
- ` + "`.AddTransition(tr, from, to)`" + ` — add transition between clips
- ` + "`.Export(profile, path)`" + ` — render output

## cuts

| Function | Description |
|----------|-------------|
| ` + "`Hard()`" + ` | Instant cut |
| ` + "`LCut(overlap)`" + ` | Audio extends past video cut |
| ` + "`JCut(overlap)`" + ` | Audio starts before video cut |
| ` + "`Dissolve(d)`" + ` | Video crossfade |
| ` + "`CrossFade(d)`" + ` | Video + audio crossfade |
| ` + "`JumpCut()`" + ` | Sharp, jarring cut |
| ` + "`DipToBlack(d)`" + ` | Fade to black and back |
| ` + "`DipToWhite(d)`" + ` | Fade to white and back |
| ` + "`Wipe(dir, d)`" + ` | Directional wipe |

## effects

| Function | Target | Description |
|----------|--------|-------------|
| ` + "`FadeIn(d)`" + ` | Video | Opacity fade in |
| ` + "`FadeOut(d)`" + ` | Video | Opacity fade out |
| ` + "`SpeedUp(f)`" + ` | Video | Playback speed |
| ` + "`SlowDown(f)`" + ` | Video | Slow motion |
| ` + "`AudioFadeIn(d)`" + ` | Audio | Volume fade in |
| ` + "`AudioFadeOut(d)`" + ` | Audio | Volume fade out |
| ` + "`Volume(v)`" + ` | Audio | Level adjustment |
| ` + "`Normalize()`" + ` | Audio | Loudness normalization |
| ` + "`AudioSpeed(f)`" + ` | Audio | Tempo change |

## export

| Preset | Resolution | Codec | Use Case |
|--------|-----------|-------|----------|
| ` + "`YouTube1080p()`" + ` | 1920x1080 | H.264 | YouTube |
| ` + "`YouTube4K()`" + ` | 3840x2160 | H.264 | YouTube 4K |
| ` + "`YouTube720p()`" + ` | 1280x720 | H.264 | YouTube 720p |
| ` + "`Reel()`" + ` | 1080x1920 | H.264 | Instagram/TikTok |
| ` + "`ProRes()`" + ` | Source | ProRes | Professional |
| ` + "`H265()`" + ` | Source | HEVC | High compression |
| ` + "`GIF()`" + ` | Source | GIF | Animated GIF |
| ` + "`MP3()`" + ` | N/A | MP3 | Audio only |
| ` + "`WAV()`" + ` | N/A | PCM | Audio only |
| ` + "`FLAC()`" + ` | N/A | FLAC | Audio only |
`
}
