# Effects

Composable effects that transform clips.

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
