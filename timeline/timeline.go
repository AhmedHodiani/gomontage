package timeline

import (
	"fmt"
	"sort"
	"time"
)

// Config holds the configuration for a new Timeline.
type Config struct {
	// Width is the output video width in pixels.
	Width int

	// Height is the output video height in pixels.
	Height int

	// FPS is the output frame rate.
	FPS float64
}

// Timeline is the top-level container that holds all tracks and clips
// for a video editing project.
type Timeline struct {
	config      Config
	videoTracks []*VideoTrack
	audioTracks []*AudioTrack
}

// New creates a new Timeline with the given configuration.
func New(cfg Config) *Timeline {
	if cfg.FPS == 0 {
		cfg.FPS = 30
	}
	return &Timeline{
		config: cfg,
	}
}

// Config returns the timeline's configuration.
func (tl *Timeline) Config() Config {
	return tl.config
}

// AddVideoTrack creates and adds a named video track to the timeline.
// Video tracks hold video clips, image clips, text overlays, and color clips.
// Tracks are layered in order — later tracks are composited on top.
func (tl *Timeline) AddVideoTrack(name string) *VideoTrack {
	track := &VideoTrack{
		name:  name,
		index: len(tl.videoTracks),
	}
	tl.videoTracks = append(tl.videoTracks, track)
	return track
}

// AddAudioTrack creates and adds a named audio track to the timeline.
// Audio tracks hold audio clips and the audio portions of video clips.
// All audio tracks are mixed together in the final output.
func (tl *Timeline) AddAudioTrack(name string) *AudioTrack {
	track := &AudioTrack{
		name:  name,
		index: len(tl.audioTracks),
	}
	tl.audioTracks = append(tl.audioTracks, track)
	return track
}

// VideoTracks returns all video tracks in order.
func (tl *Timeline) VideoTracks() []*VideoTrack {
	return tl.videoTracks
}

// AudioTracks returns all audio tracks in order.
func (tl *Timeline) AudioTracks() []*AudioTrack {
	return tl.audioTracks
}

// Duration returns the total duration of the timeline, determined by
// the latest ending clip across all tracks.
func (tl *Timeline) Duration() time.Duration {
	var maxEnd time.Duration

	for _, track := range tl.videoTracks {
		for _, entry := range track.entries {
			end := entry.StartAt + entry.Clip.Duration()
			if end > maxEnd {
				maxEnd = end
			}
		}
	}
	for _, track := range tl.audioTracks {
		for _, entry := range track.entries {
			end := entry.StartAt + entry.Clip.Duration()
			if end > maxEnd {
				maxEnd = end
			}
		}
	}

	return maxEnd
}

// Validate checks the timeline for common errors before export.
// Returns nil if everything looks good, or a descriptive error.
func (tl *Timeline) Validate() error {
	if tl.config.Width <= 0 || tl.config.Height <= 0 {
		return fmt.Errorf("invalid resolution: %dx%d", tl.config.Width, tl.config.Height)
	}
	if tl.config.FPS <= 0 {
		return fmt.Errorf("invalid FPS: %f", tl.config.FPS)
	}
	if len(tl.videoTracks) == 0 && len(tl.audioTracks) == 0 {
		return fmt.Errorf("timeline has no tracks")
	}

	totalClips := 0
	for _, track := range tl.videoTracks {
		totalClips += len(track.entries)
	}
	for _, track := range tl.audioTracks {
		totalClips += len(track.entries)
	}
	if totalClips == 0 {
		return fmt.Errorf("timeline has no clips")
	}

	// Check for overlapping clips on video tracks. A single video track
	// can only show one clip at a time — overlapping placements are an error.
	for _, track := range tl.videoTracks {
		if err := validateNoOverlap(track.Name(), track.entries); err != nil {
			return err
		}
	}

	return nil
}

// validateNoOverlap checks that no two placements on a track overlap in time.
// Entries are sorted by StartAt; if a clip's start is before the previous
// clip's end, that's an overlap.
func validateNoOverlap(trackName string, entries []Placement) error {
	if len(entries) < 2 {
		return nil
	}

	sorted := make([]Placement, len(entries))
	copy(sorted, entries)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].StartAt < sorted[j].StartAt
	})

	for i := 1; i < len(sorted); i++ {
		prevEnd := sorted[i-1].StartAt + sorted[i-1].Clip.Duration()
		currStart := sorted[i].StartAt
		if currStart < prevEnd {
			overlap := prevEnd - currStart
			return fmt.Errorf(
				"track %q: clip at %.3fs (duration %.3fs, ends %.3fs) overlaps with clip at %.3fs by %.3fs",
				trackName,
				sorted[i-1].StartAt.Seconds(),
				sorted[i-1].Clip.Duration().Seconds(),
				prevEnd.Seconds(),
				currStart.Seconds(),
				overlap.Seconds(),
			)
		}
	}

	return nil
}
