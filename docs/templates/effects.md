# Effects

Composable effects that transform clips.

## Applying Effects to Clips

Use the `WithEffect` method on any clip type to apply effects. Effects are
composable — call `WithEffect` multiple times to stack them. Like all clip
transforms, `WithEffect` returns a new clip and never mutates the original.

```go
// Single effect
interview := clip.NewVideo("interview.mp4").
    WithEffect(effects.SpeedUp(2.0))

// Stack multiple effects
intro := clip.NewVideo("intro.mp4").
    Trim(0, 10*time.Second).
    WithEffect(effects.FadeIn(1 * time.Second)).
    WithEffect(effects.FadeOut(2 * time.Second)).
    WithEffect(effects.SpeedUp(1.5))

// Works on all clip types
music := clip.NewAudio("bgm.mp3").
    WithEffect(effects.Volume(0.3)).
    WithEffect(effects.AudioFadeIn(2 * time.Second)).
    WithEffect(effects.Normalize())

logo := clip.NewImage("logo.png").
    WithDuration(5 * time.Second).
    WithEffect(effects.FadeIn(1 * time.Second))

title := clip.NewText("Chapter 1", clip.DefaultTextStyle()).
    WithDuration(4 * time.Second).
    WithEffect(effects.FadeIn(1 * time.Second)).
    WithEffect(effects.FadeOut(1 * time.Second))
```

You can read back applied effects with the `Effects()` method:

```go
c := clip.NewVideo("test.mp4").
    WithEffect(effects.FadeIn(1 * time.Second)).
    WithEffect(effects.SpeedUp(2.0))

for _, e := range c.Effects() {
    fmt.Printf("%s -> %s\n", e.Name(), e.FilterName())
}
// Output:
// fade_in -> fade
// speed -> setpts
```

## Video Effects

### Fade In / Fade Out
```go
effects.FadeIn(1*time.Second)
effects.FadeOut(2*time.Second)
effects.FadeOutAt(10*time.Second, 2*time.Second) // Start at specific time
```

### Speed
```go
effects.SpeedUp(2.0)   // 2x speed (timelapse)
effects.SpeedUp(0.5)   // Half speed (slow motion)
effects.SlowDown(2.0)  // Same as SpeedUp(0.5)
```

## Audio Effects

### Audio Fade
```go
effects.AudioFadeIn(3*time.Second)
effects.AudioFadeOut(5*time.Second)
```

### Volume
```go
effects.Volume(0.5)  // Half volume
effects.Volume(0.0)  // Mute
effects.Volume(2.0)  // Double volume
```

### Normalize
```go
effects.Normalize()          // Target -16 LUFS (broadcast standard)
effects.NormalizeTo(-23.0)   // Custom LUFS target
```

### Audio Speed
```go
effects.AudioSpeed(1.5)  // 1.5x playback speed
```
