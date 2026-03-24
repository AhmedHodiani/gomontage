package clip

import "time"

// SubRange returns a new Clip containing only the portion visible within the
// given window [start, end) in clip-local time (relative to the clip's start,
// not the timeline). The returned clip's Duration equals end - start.
//
// For file-backed clips (video, audio), the source trim window is adjusted so
// FFmpeg reads only the needed frames. For generated clips (color, text, image),
// the duration is shortened.
//
// start and end are clamped to [0, clip.Duration()]. If the window is empty
// (start >= end after clamping), nil is returned.
//
// Effects, volume, fades, and other properties are preserved on the returned
// clip.
func SubRange(c Clip, start, end time.Duration) Clip {
	dur := c.Duration()

	// Clamp to valid range.
	if start < 0 {
		start = 0
	}
	if end > dur {
		end = dur
	}
	if start >= end {
		return nil
	}

	// If the window covers the full clip, return as-is (no work needed).
	if start == 0 && end == dur {
		return c
	}

	switch v := c.(type) {
	case *VideoClip:
		return subRangeVideo(v, start, end)
	case *AudioClip:
		return subRangeAudio(v, start, end)
	case *ImageClip:
		return subRangeImage(v, end-start)
	case *ColorClip:
		return subRangeColor(v, end-start)
	case *TextClip:
		return subRangeText(v, end-start)
	default:
		return nil
	}
}

// subRangeVideo trims a video clip to the given clip-local window.
// It translates the local window into source-file coordinates.
func subRangeVideo(c *VideoClip, start, end time.Duration) *VideoClip {
	n := &VideoClip{Base: *c.base()}

	// Translate clip-local time to source-file time.
	// For clips with speed effects, we need the source-time ratio.
	// TrimStart/TrimEnd are always in source coordinates.
	// Duration() accounts for speed effects.
	//
	// sourceWindow = TrimEnd - TrimStart (always, regardless of speed)
	// localDuration = Duration() (accounts for speed)
	// ratio = sourceWindow / localDuration
	//
	// So to convert local time to source time:
	//   sourceTime = TrimStart + localTime * ratio
	localDur := c.Duration()
	sourceStart := c.TrimStart()
	sourceEnd := c.TrimEnd()
	sourceWindow := sourceEnd - sourceStart

	if localDur > 0 && sourceWindow > 0 {
		ratio := float64(sourceWindow) / float64(localDur)
		n.trimStart = sourceStart + time.Duration(float64(start)*ratio)
		n.trimEnd = sourceStart + time.Duration(float64(end)*ratio)
	} else {
		n.trimStart = sourceStart + start
		n.trimEnd = sourceStart + end
	}

	n.duration = end - start
	n.trimmed = true

	return n
}

// subRangeAudio trims an audio clip to the given clip-local window.
func subRangeAudio(c *AudioClip, start, end time.Duration) *AudioClip {
	n := &AudioClip{Base: *c.base()}

	localDur := c.Duration()
	sourceStart := c.TrimStart()
	sourceEnd := c.TrimEnd()
	sourceWindow := sourceEnd - sourceStart

	if localDur > 0 && sourceWindow > 0 {
		ratio := float64(sourceWindow) / float64(localDur)
		n.trimStart = sourceStart + time.Duration(float64(start)*ratio)
		n.trimEnd = sourceStart + time.Duration(float64(end)*ratio)
	} else {
		n.trimStart = sourceStart + start
		n.trimEnd = sourceStart + end
	}

	n.duration = end - start
	n.trimmed = true

	return n
}

// subRangeImage shortens an image clip to the given duration.
func subRangeImage(c *ImageClip, dur time.Duration) *ImageClip {
	n := &ImageClip{Base: *c.base()}
	n.duration = dur
	n.trimEnd = dur
	return n
}

// subRangeColor shortens a color clip to the given duration.
func subRangeColor(c *ColorClip, dur time.Duration) *ColorClip {
	n := &ColorClip{Base: *c.base(), Color: c.Color}
	n.duration = dur
	n.trimEnd = dur
	return n
}

// subRangeText shortens a text clip to the given duration.
func subRangeText(c *TextClip, dur time.Duration) *TextClip {
	n := &TextClip{Base: *c.base(), Text: c.Text, Style: c.Style}
	n.duration = dur
	n.trimEnd = dur
	return n
}
