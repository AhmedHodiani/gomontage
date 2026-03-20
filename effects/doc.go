// Package effects provides composable audio and video effects for Gomontage.
//
// Effects are applied to clips via the clip's WithEffect method or directly
// referenced by the timeline compiler. They represent transformations that
// map to FFmpeg filters.
//
// # Video Effects
//
//	effects.FadeIn(1 * time.Second)
//	effects.FadeOut(2 * time.Second)
//	effects.SpeedUp(2.0)   // 2x speed
//	effects.SlowDown(0.5)  // half speed
//
// # Audio Effects
//
//	effects.AudioFadeIn(1 * time.Second)
//	effects.AudioFadeOut(2 * time.Second)
//	effects.Volume(0.5)  // half volume
//	effects.Normalize()  // normalize audio levels
package effects
