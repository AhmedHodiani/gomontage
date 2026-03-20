// Package cuts provides transition types for the Gomontage timeline.
//
// Cuts define how two adjacent clips blend together. They range from simple
// hard cuts (instant switch) to complex transitions like L-cuts and J-cuts
// that manipulate the timing between audio and video independently.
//
// # Cut Types
//
// Hard cut: Instant switch from one clip to the next (default).
//
//	cuts.Hard()
//
// L-cut: Audio from the first clip continues playing over the second clip's video.
// Common in interviews where you cut to B-roll while the speaker keeps talking.
//
//	cuts.LCut(2 * time.Second) // Audio extends 2s into next clip
//
// J-cut: Audio from the second clip starts playing before its video appears.
// Builds anticipation — you hear the next scene before you see it.
//
//	cuts.JCut(1 * time.Second) // Audio starts 1s before video
//
// Dissolve: Crossfade between two video clips.
//
//	cuts.Dissolve(1 * time.Second)
//
// CrossFade: Crossfade both video and audio simultaneously.
//
//	cuts.CrossFade(1 * time.Second)
//
// Jump cut: Removes a section and snaps clips together with no transition.
//
//	cuts.JumpCut()
//
// Dip to black/white: Fades out to a color, then fades back in.
//
//	cuts.DipToBlack(1 * time.Second)
//	cuts.DipToWhite(500 * time.Millisecond)
package cuts
