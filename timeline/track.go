package timeline

import (
	"time"

	"github.com/ahmedhodiani/gomontage/clip"
)

// TrackType identifies whether a track holds video or audio.
type TrackType int

const (
	// TrackTypeVideo holds video clips composited as visual layers.
	TrackTypeVideo TrackType = iota
	// TrackTypeAudio holds audio clips mixed together.
	TrackTypeAudio
)

// Placement specifies where a clip starts on the timeline.
type Placement struct {
	// StartAt is the absolute time on the timeline where the clip begins.
	StartAt time.Duration

	// Clip is the media clip placed at this position.
	Clip clip.Clip
}

// At creates a Placement that starts a clip at the given time on the timeline.
//
// Example:
//
//	track.Add(myClip, timeline.At(5 * time.Second))
func At(t time.Duration) time.Duration {
	return t
}

// VideoTrack is a named track that holds visual clips (video, image, text, color).
// Clips are placed at specific times and rendered in placement order for the
// same time position.
type VideoTrack struct {
	name    string
	index   int
	entries []Placement
}

// Name returns the track's name.
func (t *VideoTrack) Name() string {
	return t.name
}

// Index returns the track's index (0 = bottom layer, higher = on top).
func (t *VideoTrack) Index() int {
	return t.index
}

// Entries returns all clip placements on this track, in the order they were added.
func (t *VideoTrack) Entries() []Placement {
	return t.entries
}

// Add places a clip on this track at the specified start time.
//
// Example:
//
//	track.Add(interview.Trim(0, 10*time.Second), timeline.At(0))
//	track.Add(broll.Trim(5*time.Second, 15*time.Second), timeline.At(10*time.Second))
func (t *VideoTrack) Add(c clip.Clip, startAt time.Duration) {
	t.entries = append(t.entries, Placement{
		StartAt: startAt,
		Clip:    c,
	})
}

// AddSequence places multiple clips back-to-back on this track.
// The first clip starts at the current track end (or 0 if empty).
//
// Example:
//
//	track.AddSequence(intro, mainContent, outro)
func (t *VideoTrack) AddSequence(clips ...clip.Clip) {
	startAt := t.End()
	for _, c := range clips {
		t.entries = append(t.entries, Placement{
			StartAt: startAt,
			Clip:    c,
		})
		startAt += c.Duration()
	}
}

// End returns the time at which the last clip on this track ends.
// Returns 0 if the track is empty.
func (t *VideoTrack) End() time.Duration {
	var maxEnd time.Duration
	for _, entry := range t.entries {
		end := entry.StartAt + entry.Clip.Duration()
		if end > maxEnd {
			maxEnd = end
		}
	}
	return maxEnd
}

// TransitionAll sets the same transition between every pair of adjacent clips
// on this track (in placement order). This is a convenience for applying
// uniform transitions.
//
// Example:
//
//	track.AddSequence(clip1, clip2, clip3)
//	track.TransitionAll(cuts.Dissolve(1 * time.Second))
func (t *VideoTrack) TransitionAll(tr Transition) []TransitionEntry {
	var entries []TransitionEntry
	for i := 0; i < len(t.entries)-1; i++ {
		entries = append(entries, TransitionEntry{
			Transition: tr,
			FromClip:   t.entries[i].Clip,
			ToClip:     t.entries[i+1].Clip,
		})
	}
	return entries
}

// AudioTrack is a named track that holds audio clips.
// All audio tracks are mixed together in the final output.
type AudioTrack struct {
	name    string
	index   int
	entries []Placement
}

// Name returns the track's name.
func (t *AudioTrack) Name() string {
	return t.name
}

// Index returns the track's index.
func (t *AudioTrack) Index() int {
	return t.index
}

// Entries returns all clip placements on this track.
func (t *AudioTrack) Entries() []Placement {
	return t.entries
}

// Add places an audio clip on this track at the specified start time.
//
// Example:
//
//	voiceover.Add(narration, timeline.At(0))
//	music.Add(bgm.WithVolume(0.3), timeline.At(0))
func (t *AudioTrack) Add(c clip.Clip, startAt time.Duration) {
	t.entries = append(t.entries, Placement{
		StartAt: startAt,
		Clip:    c,
	})
}

// AddSequence places multiple audio clips back-to-back on this track.
//
// Example:
//
//	voice.AddSequence(intro, chapter1, chapter2, outro)
func (t *AudioTrack) AddSequence(clips ...clip.Clip) {
	startAt := t.End()
	for _, c := range clips {
		t.entries = append(t.entries, Placement{
			StartAt: startAt,
			Clip:    c,
		})
		startAt += c.Duration()
	}
}

// End returns the time at which the last clip on this track ends.
func (t *AudioTrack) End() time.Duration {
	var maxEnd time.Duration
	for _, entry := range t.entries {
		end := entry.StartAt + entry.Clip.Duration()
		if end > maxEnd {
			maxEnd = end
		}
	}
	return maxEnd
}
