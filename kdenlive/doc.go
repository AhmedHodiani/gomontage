// Package kdenlive provides functionality to export GoMontage timelines as
// .kdenlive project files that can be opened in the Kdenlive video editor.
//
// This enables a workflow where you programmatically define your video edit using
// GoMontage's timeline model, then open the project in Kdenlive's GUI editor
// for further manual refinement and rendering.
//
// Only file-backed clips (VideoClip, AudioClip, ImageClip) are currently
// supported. Generated clips (ColorClip, TextClip) will return an error.
//
// # Quick Start
//
//	tl := gomontage.NewTimeline(gomontage.HD())
//	video := tl.AddVideoTrack("main")
//	video.Add(clip.NewVideo("intro.mp4"), gomontage.At(0))
//	audio := tl.AddAudioTrack("narration")
//	audio.Add(clip.NewAudio("speech.mp3"), gomontage.At(0))
//
//	err := kdenlive.Export(tl, "project.kdenlive")
//	if err != nil {
//	    log.Fatal(err)
//	}
package kdenlive