// Package timeline provides the track-based timeline composition model.
//
// A Timeline is the top-level container for a Gomontage project. It holds
// named tracks (video and audio) where clips are placed at specific times.
// When you call Export, the timeline is compiled into an FFmpeg filter graph
// and rendered to a file.
//
// # Creating a Timeline
//
//	tl := timeline.New(timeline.Config{
//	    Width:  1920,
//	    Height: 1080,
//	    FPS:    30,
//	})
//
// # Working with Tracks
//
//	video := tl.AddVideoTrack("main")
//	audio := tl.AddAudioTrack("narration")
//
//	video.Add(myClip, timeline.At(0))
//	video.AddSequence(clip1, clip2, clip3)
//
// # Exporting
//
//	tl.Export(profile, "output/final.mp4")
package timeline
