package timeline

import (
	"time"

	"github.com/ahmedhodiani/gomontage/clip"
)

// SubRange returns a new Timeline containing only the portion of the original
// between start and end. Clips entirely outside the range are excluded, clips
// partially overlapping are trimmed to fit, and all StartAt values are shifted
// so the range begins at t=0.
//
// This is useful for quickly exporting a preview of a specific section without
// rendering the entire timeline.
//
// Example:
//
//	// Export only the first 2 minutes of a 1-hour timeline.
//	preview := tl.SubRange(0, 2*time.Minute)
//	preview.Export(export.YouTube1080p(), "preview.mp4")
func (tl *Timeline) SubRange(start, end time.Duration) *Timeline {
	if start < 0 {
		start = 0
	}
	if end > tl.Duration() {
		end = tl.Duration()
	}
	if start >= end {
		// Empty range — return a timeline with no clips.
		return New(tl.config)
	}

	sub := New(tl.config)

	// Filter video tracks.
	for _, track := range tl.videoTracks {
		subTrack := sub.AddVideoTrack(track.Name())
		filterPlacements(track.Entries(), start, end, subTrack)
	}

	// Filter audio tracks.
	for _, track := range tl.audioTracks {
		subTrack := sub.AddAudioTrack(track.Name())
		filterPlacements(track.Entries(), start, end, subTrack)
	}

	return sub
}

// trackAdder is satisfied by both VideoTrack and AudioTrack.
type trackAdder interface {
	Add(c clip.Clip, startAt time.Duration)
}

// filterPlacements examines each placement against the [rangeStart, rangeEnd)
// window and adds the visible portion to the target track with a shifted StartAt.
func filterPlacements(entries []Placement, rangeStart, rangeEnd time.Duration, target trackAdder) {
	for _, entry := range entries {
		clipStart := entry.StartAt
		clipEnd := entry.StartAt + entry.Clip.Duration()

		// Entirely outside the range — skip.
		if clipEnd <= rangeStart || clipStart >= rangeEnd {
			continue
		}

		// Calculate the visible portion in clip-local time.
		visibleStart := time.Duration(0)
		if clipStart < rangeStart {
			visibleStart = rangeStart - clipStart
		}

		visibleEnd := entry.Clip.Duration()
		if clipEnd > rangeEnd {
			visibleEnd = entry.Clip.Duration() - (clipEnd - rangeEnd)
		}

		// SubRange the clip to the visible window.
		trimmed := clip.SubRange(entry.Clip, visibleStart, visibleEnd)
		if trimmed == nil {
			continue
		}

		// Shift StartAt so the range starts at t=0.
		newStart := clipStart - rangeStart
		if newStart < 0 {
			newStart = 0
		}

		target.Add(trimmed, newStart)
	}
}
