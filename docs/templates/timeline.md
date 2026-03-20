# Timeline & Tracks

The Timeline is the master container for your video project. It holds named
tracks where clips are placed at specific time positions.

## Creating a Timeline

```go
tl := timeline.New(timeline.Config{
    Width:  1920,
    Height: 1080,
    FPS:    30,
})
```

## Tracks

### Video Tracks

Video tracks hold visual clips. Multiple tracks are composited (layered)
in order — later tracks appear on top.

```go
main := tl.AddVideoTrack("main")       // Base layer
overlay := tl.AddVideoTrack("overlay")  // On top
```

### Audio Tracks

Audio tracks hold audio clips. All audio tracks are mixed together.

```go
voice := tl.AddAudioTrack("narration")
music := tl.AddAudioTrack("music")
sfx := tl.AddAudioTrack("sfx")
```

## Placing Clips

### At a Specific Time

```go
track.Add(myClip, timeline.At(5*time.Second))
```

### Sequentially (Back-to-Back)

```go
track.AddSequence(intro, mainContent, outro)
// intro starts at 0, mainContent starts when intro ends, etc.
```

## Timeline Duration

Duration is automatically calculated from the latest-ending clip:

```go
duration := tl.Duration() // Returns the total timeline length
```

## Validation

Check for errors before exporting:

```go
if err := tl.Validate(); err != nil {
    log.Fatal(err)
}
```
