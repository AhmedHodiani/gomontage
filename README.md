# GoMontage

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go&logoColor=white)
![FFmpeg](https://img.shields.io/badge/FFmpeg-required-007808?style=flat&logo=ffmpeg&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-green?style=flat)

> GoMontage is a programmatic video editing framework for Go. Instead of reaching for a GUI editor, you describe your edit in code — trim clips, arrange tracks, layer audio, apply effects — and GoMontage compiles it into an FFmpeg command and renders it. If you can write Go, you can automate video production.

![titlecard](images/titlecard.png)

---

## How GoMontage Works

GoMontage has four layers that take your code all the way to a rendered video file:

1. **Clips** — Load media with `clip.NewVideo()`, `clip.NewAudio()`, etc. Clips are immutable; every transform (`.Trim()`, `.WithVolume()`) returns a new instance.
2. **Timeline** — Add clips to named video and audio tracks using `tl.AddVideoTrack()`. Positioning is explicit: `timeline.At(5 * time.Second)` places a clip at the 5-second mark.
3. **Compiler** — When you call `tl.Export()`, the Compiler walks the timeline and builds an FFmpeg filter graph — handling scaling, mixing, sequencing, and effects automatically.
4. **Engine** — The filter graph is handed to the Engine, which constructs the final FFmpeg command and runs it.

You write Go. GoMontage handles the FFmpeg.

---

## Requirements

- Go 1.25+
- FFmpeg & ffprobe installed and available in `PATH`

---

## Install

```bash
go install github.com/ahmedhodiani/gomontage/cmd/gomontage@latest
```

---

## Quick Start

### 1. Scaffold a project

```bash
gomontage init my-video
cd my-video
```

This creates the project structure with a starter `main.go`, config file, and folders for your media files.

### 2. Add your media

Drop video, audio, and image files into `resources/video/`, `resources/audio/`, etc.

### 3. Edit `main.go`

```go
package main

import (
    "time"

    "github.com/ahmedhodiani/gomontage/clip"
    "github.com/ahmedhodiani/gomontage/export"
    "github.com/ahmedhodiani/gomontage/timeline"
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

### 4. Render

```bash
gomontage run
```

---

## CLI Commands

| Command | Description |
|---|---|
| `gomontage init <name>` | Scaffold a new project |
| `gomontage run` | Run `main.go` to render the video |
| `gomontage probe <file>` | Inspect media file metadata |
| `gomontage validate` | Validate the project config |
| `gomontage docs` | Generate API documentation into `docs/` |

### Generate docs

```bash
gomontage docs
```

Writes a full markdown API reference and guides into the `docs/` directory.
