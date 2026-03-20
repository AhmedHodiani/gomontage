// Package clip provides media clip types for Gomontage.
//
// Clips are the fundamental building blocks of a Gomontage project. Each clip
// represents a piece of media — a video file, audio file, image, text overlay,
// or solid color. Clips are immutable: every transformation method returns a
// new clip, leaving the original unchanged.
//
// # Loading Clips
//
//	video := clip.NewVideo("interview.mp4")
//	audio := clip.NewAudio("narration.wav")
//	title := clip.NewText("Hello World", clip.TextStyle{Size: 72})
//	bg    := clip.NewColor("#000000", 1920, 1080)
//	img   := clip.NewImage("logo.png")
//
// # Transforming Clips
//
// All transforms return new clips:
//
//	trimmed := video.Trim(5*time.Second, 15*time.Second)
//	quiet   := audio.WithVolume(0.3)
//	faded   := video.WithFadeIn(1*time.Second).WithFadeOut(2*time.Second)
package clip
