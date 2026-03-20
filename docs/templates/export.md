# Export Profiles

Export profiles configure the output format, codec, quality, and resolution.

## Built-in Presets

### YouTube
```go
export.YouTube1080p()  // 1920x1080, H.264, AAC, CRF 18
export.YouTube4K()     // 3840x2160, H.264, AAC, CRF 18
export.YouTube720p()   // 1280x720, H.264, AAC, CRF 20
```

### Social Media
```go
export.Reel()  // 1080x1920 (9:16 vertical), H.264
```

### Professional
```go
export.ProRes()  // Apple ProRes 422, MOV container
export.H265()    // H.265/HEVC, better compression
```

### Audio Only
```go
export.MP3()   // MP3, 192kbps
export.WAV()   // Uncompressed PCM
export.FLAC()  // Lossless compressed
```

### Animated
```go
export.GIF()  // Animated GIF
```

## Custom Profiles

Use the fluent builder for full control:

```go
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
```
