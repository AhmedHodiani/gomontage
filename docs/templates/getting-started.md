# Getting Started with Gomontage

Gomontage is a programmatic video editing framework for Go. Write Go code
to edit videos instead of using a GUI.

## Installation

```bash
go install github.com/ahmedhodiani/gomontage/cmd/gomontage@latest
```

## Create a New Project

```bash
gomontage init my-video
cd my-video
```

This creates the following structure:

```
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
```

## Your First Edit

Edit `main.go` to describe your video:

```go
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
```

## Run Your Project

```bash
gomontage run
```

## Inspect Media Files

```bash
gomontage probe resources/video/footage.mp4
```

## Key Concepts

1. **Clips** are your media building blocks (video, audio, image, text, color)
2. **Tracks** are named layers on the timeline (video tracks, audio tracks)
3. **Timeline** is the master container that holds all tracks
4. **Effects** are transformations applied to clips
5. **Export profiles** configure the output format and quality
