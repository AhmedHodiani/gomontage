// Package export provides output profiles and presets for Gomontage.
//
// Export profiles encapsulate all the FFmpeg output settings (codec, bitrate,
// pixel format, audio codec, sample rate, etc.) into named presets. Users
// select a profile when exporting their timeline.
//
// # Built-in Presets
//
//	export.YouTube1080p()  // 1920x1080, H.264, AAC, optimized for YouTube
//	export.YouTube4K()     // 3840x2160, H.264, AAC
//	export.Reel()          // 1080x1920, vertical, H.264 for Instagram/TikTok
//	export.ProRes()        // Apple ProRes 422 for professional workflows
//	export.MP3()           // Audio-only MP3 export
//	export.WAV()           // Audio-only WAV export
//
// # Custom Profiles
//
//	custom := export.NewProfile().
//	    WithCodec("libx265").
//	    WithBitrate("8M").
//	    WithAudioCodec("aac").
//	    WithAudioBitrate("320k").
//	    Build()
package export
