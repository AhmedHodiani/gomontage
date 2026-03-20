# Clips

Clips are the fundamental building blocks of a Gomontage project. Every clip
is **immutable** — transform methods return new clips, leaving originals unchanged.

## Clip Types

### VideoClip

Load video from a file. Contains both video and audio streams.

```go
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
```

### AudioClip

Load audio from a file (WAV, MP3, FLAC, etc.).

```go
narration := clip.NewAudio("resources/audio/narration.wav")
music := clip.NewAudio("resources/audio/bgm.mp3").WithVolume(0.3)

// Trim
intro := narration.Trim(0, 30*time.Second)

// Fade in/out
faded := music.WithFadeIn(3*time.Second).WithFadeOut(5*time.Second)
```

### ImageClip

Static images used as video clips. Default duration is 5 seconds.

```go
logo := clip.NewImage("resources/images/logo.png").WithDuration(3*time.Second)
bg := clip.NewImage("resources/images/background.jpg").WithDuration(10*time.Second)
```

### TextClip

Dynamically rendered text overlays.

```go
title := clip.NewText("Chapter 1", clip.TextStyle{
    Font:  "resources/fonts/bold.ttf",
    Size:  72,
    Color: "#FFFFFF",
}).WithDuration(4*time.Second).WithFadeIn(500*time.Millisecond)
```

### ColorClip

Solid color rectangles for backgrounds.

```go
black := clip.NewColor("#000000", 1920, 1080).WithDuration(2*time.Second)
```

## Immutability

All transforms return new clips. The original is never modified:

```go
original := clip.NewVideo("video.mp4")
trimmed := original.Trim(0, 10*time.Second)
// original is still the full video
// trimmed is a new clip with just the first 10 seconds
```

## Chaining

Transforms can be chained fluently:

```go
result := clip.NewVideo("raw.mp4").
    Trim(5*time.Second, 25*time.Second).
    WithVolume(0.8).
    WithFadeIn(1*time.Second).
    WithFadeOut(2*time.Second)
```
