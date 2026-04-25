package kdenlive

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ahmedhodiani/gomontage/clip"
	"github.com/ahmedhodiani/gomontage/timeline"
)

// Export writes a .kdenlive project file from a GoMontage Timeline.
// Only file-backed clips (VideoClip, AudioClip, ImageClip) are supported.
func Export(tl *timeline.Timeline, outputPath string) error {
	data, err := ExportBytes(tl)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, data, 0644)
}

// ExportBytes returns the kdenlive project file content as a byte slice.
func ExportBytes(tl *timeline.Timeline) ([]byte, error) {
	if err := validateTimeline(tl); err != nil {
		return nil, err
	}
	b := newBuilder(tl)
	return b.build(), nil
}

func validateTimeline(tl *timeline.Timeline) error {
	for _, vt := range tl.VideoTracks() {
		for _, e := range vt.Entries() {
			if e.Clip.SourcePath() == "" {
				return fmt.Errorf("kdenlive: unsupported generated clip on video track %q — only file-backed clips are supported", vt.Name())
			}
		}
	}
	for _, at := range tl.AudioTracks() {
		for _, e := range at.Entries() {
			if e.Clip.SourcePath() == "" {
				return fmt.Errorf("kdenlive: unsupported generated clip on audio track %q — only file-backed clips are supported", at.Name())
			}
		}
	}
	return nil
}

type trackKind int

const (
	trackKindVideo trackKind = iota
	trackKindAudio
	trackKindAudioCompanion
)

type chainInfo struct {
	id         string
	sourcePath string
}

type trackInfo struct {
	kind               trackKind
	tractorID          string
	playlist1ID        string
	playlist2ID        string
	entries            []trackEntry
	videoTrackEntries  []trackEntry
}

type trackEntry struct {
	startAt time.Duration
	c       clip.Clip
}

type unmutedRange struct {
	start time.Duration
	end   time.Duration
}

type builder struct {
	tl *timeline.Timeline

	chainCount  int
	filterCount int
	trackCount int
	transCount  int

	chains    map[string]string // sourcePath -> chainID
	chainList []chainInfo

	uuid  string
	docID string

	tracks        []trackInfo
	unmutedRanges []unmutedRange // times where unmuted video clips have audio

	buf bytes.Buffer
	ind  int
}

func newBuilder(tl *timeline.Timeline) *builder {
	return &builder{
		tl:     tl,
		chains: make(map[string]string),
		uuid:   generateUUID(),
		docID:  fmt.Sprintf("%d", time.Now().UnixMilli()),
	}
}

func (b *builder) nextChainID() string {
	id := fmt.Sprintf("chain%d", b.chainCount)
	b.chainCount++
	return id
}

func (b *builder) nextFilterID() string {
	id := fmt.Sprintf("filter%d", b.filterCount)
	b.filterCount++
	return id
}

func (b *builder) nextTrackID() int {
	id := b.trackCount
	b.trackCount++
	return id
}

func (b *builder) nextTransID() string {
	id := fmt.Sprintf("transition%d", b.transCount)
	b.transCount++
	return id
}

func (b *builder) getOrCreateChain(sourcePath string) string {
	absPath, err := filepath.Abs(sourcePath)
	if err != nil {
		absPath = sourcePath
	}
	if id, ok := b.chains[absPath]; ok {
		return id
	}
	id := b.nextChainID()
	b.chains[absPath] = id
	b.chainList = append(b.chainList, chainInfo{id: id, sourcePath: absPath})
	return id
}

func (b *builder) chainIDForClip(c clip.Clip) string {
	absPath, err := filepath.Abs(c.SourcePath())
	if err != nil {
		absPath = c.SourcePath()
	}
	return b.chains[absPath]
}

func (b *builder) indent() string {
	return strings.Repeat("    ", b.ind)
}

func (b *builder) writeLine(format string, args ...interface{}) {
	b.buf.WriteString(b.indent())
	b.buf.WriteString(fmt.Sprintf(format, args...))
	b.buf.WriteByte('\n')
}

func writeAttr(name, value string) string {
	return fmt.Sprintf(`%s="%s"`, name, escapeAttr(value))
}

func (b *builder) writeOpenTag(tag string, attrs ...string) {
	b.writeLine("<%s%s>", tag, joinAttrs(attrs))
	b.ind++
}

func (b *builder) writeCloseTag(tag string) {
	b.ind--
	b.writeLine("</%s>", tag)
}

func (b *builder) writeEmptyTag(tag string, attrs ...string) {
	b.writeLine("<%s/>", joinAttrs2(tag, attrs))
}

func (b *builder) writeProperty(name, value string) {
	b.writeLine(`<property name="%s">%s</property>`, escapeAttr(name), escapeText(value))
}

func (b *builder) build() []byte {
	cfg := b.tl.Config()

	for _, vt := range b.tl.VideoTracks() {
		entries := make([]trackEntry, len(vt.Entries()))
		for i, e := range vt.Entries() {
			entries[i] = trackEntry{startAt: e.StartAt, c: e.Clip}
		}
		sort.SliceStable(entries, func(i, j int) bool {
			return entries[i].startAt < entries[j].startAt
		})
		idx := b.nextTrackID()
		b.tracks = append(b.tracks, trackInfo{
			kind:        trackKindVideo,
			tractorID:   fmt.Sprintf("tractor%d", idx),
			playlist1ID: fmt.Sprintf("playlist%d", idx*2),
			playlist2ID: fmt.Sprintf("playlist%d", idx*2+1),
			entries:    entries,
		})

		hasUnmuted := false
		for _, e := range entries {
			if e.c.SourcePath() != "" && e.c.HasAudio() {
				hasUnmuted = true
				b.unmutedRanges = append(b.unmutedRanges, unmutedRange{
					start: e.startAt,
					end:   e.startAt + e.c.Duration(),
				})
			}
		}
		if hasUnmuted {
			idx := b.nextTrackID()
			audioEntries := make([]trackEntry, 0, len(entries))
			for _, e := range entries {
				if e.c.SourcePath() != "" && e.c.HasAudio() {
					audioEntries = append(audioEntries, e)
				}
			}
			b.tracks = append(b.tracks, trackInfo{
				kind:             trackKindAudioCompanion,
				tractorID:        fmt.Sprintf("tractor%d", idx),
				playlist1ID:      fmt.Sprintf("playlist%d", idx*2),
				playlist2ID:      fmt.Sprintf("playlist%d", idx*2+1),
				entries:          audioEntries,
				videoTrackEntries: entries,
			})
		}
	}

	for _, at := range b.tl.AudioTracks() {
		entries := make([]trackEntry, len(at.Entries()))
		for i, e := range at.Entries() {
			entries[i] = trackEntry{startAt: e.StartAt, c: e.Clip}
		}
		sort.SliceStable(entries, func(i, j int) bool {
			return entries[i].startAt < entries[j].startAt
		})
		idx := b.nextTrackID()
		b.tracks = append(b.tracks, trackInfo{
			kind:        trackKindAudio,
			tractorID:   fmt.Sprintf("tractor%d", idx),
			playlist1ID: fmt.Sprintf("playlist%d", idx*2),
			playlist2ID: fmt.Sprintf("playlist%d", idx*2+1),
			entries:    entries,
		})
	}

	for i := range b.tracks {
		for _, e := range b.tracks[i].entries {
			if e.c.SourcePath() != "" {
				b.getOrCreateChain(e.c.SourcePath())
			}
		}
	}

	timelineDur := b.tl.Duration()
	if timelineDur < 5*time.Minute {
		timelineDur = 5 * time.Minute
	}

	b.writeLine(`<?xml version='1.0' encoding='utf-8'?>`)
	b.writeOpenTag("mlt",
		writeAttr("LC_NUMERIC", "en_US.UTF-8"),
		writeAttr("producer", "main_bin"),
		writeAttr("version", "7.25.0"),
	)
	b.writeProfile(cfg)
	b.writeBlackProducer(timelineDur)
	b.writeChains()
	b.writeTracks()
	b.writeTimelineTractor(timelineDur)
	b.writeMainBin()
	b.writeFinalTractor()
	b.writeCloseTag("mlt")

	return b.buf.Bytes()
}

func (b *builder) writeProfile(cfg timeline.Config) {
	fpsDen := 1
	fpsNum := int(math.Round(cfg.FPS))
	if cfg.FPS != float64(fpsNum) {
		fpsNum = int(math.Round(cfg.FPS * 1001))
		fpsDen = 1001
	}

	desc := fmt.Sprintf("%dx%d, %g fps", cfg.Width, cfg.Height, cfg.FPS)
	darNum, darDen := 16, 9
	if cfg.Width > 0 && cfg.Height > 0 {
		g := gcd(cfg.Width, cfg.Height)
		w, h := cfg.Width/g, cfg.Height/g
		if w <= 100 && h <= 100 {
			darNum, darDen = w, h
		}
	}

	b.writeEmptyTag("profile",
		writeAttr("colorspace", "709"),
		writeAttr("description", desc),
		writeAttr("display_aspect_den", fmt.Sprintf("%d", darDen)),
		writeAttr("display_aspect_num", fmt.Sprintf("%d", darNum)),
		writeAttr("frame_rate_den", fmt.Sprintf("%d", fpsDen)),
		writeAttr("frame_rate_num", fmt.Sprintf("%d", fpsNum)),
		writeAttr("height", fmt.Sprintf("%d", cfg.Height)),
		writeAttr("progressive", "1"),
		writeAttr("sample_aspect_den", "1"),
		writeAttr("sample_aspect_num", "1"),
		writeAttr("width", fmt.Sprintf("%d", cfg.Width)),
	)
}

func (b *builder) writeBlackProducer(dur time.Duration) {
	b.writeOpenTag("producer",
		writeAttr("id", "producer0"),
		writeAttr("in", "00:00:00.000"),
		writeAttr("out", formatTime(dur)),
	)
	b.writeProperty("length", "2147483647")
	b.writeProperty("eof", "continue")
	b.writeProperty("resource", "black")
	b.writeProperty("aspect_ratio", "1")
	b.writeProperty("mlt_service", "color")
	b.writeProperty("kdenlive:playlistid", "black_track")
	b.writeProperty("mlt_image_format", "rgba")
	b.writeProperty("set.test_audio", "0")
	b.writeCloseTag("producer")
}

func (b *builder) writeChains() {
	for _, ci := range b.chainList {
		b.writeOpenTag("chain", writeAttr("id", ci.id))
		b.writeProperty("resource", ci.sourcePath)
		b.writeCloseTag("chain")
	}
}

func (b *builder) writeTracks() {
	for i := range b.tracks {
		ti := &b.tracks[i]
		switch ti.kind {
		case trackKindVideo:
			b.writeContentPlaylist(ti, ti.playlist1ID, allContent)
			b.writeEmptyPlaylist(ti.playlist2ID, ti.kind)
		case trackKindAudio:
			b.writeContentPlaylist(ti, ti.playlist1ID, allContent)
			b.writeEmptyPlaylist(ti.playlist2ID, ti.kind)
		case trackKindAudioCompanion:
			b.writeAudioCompanionPlaylist(ti)
			b.writeEmptyPlaylist(ti.playlist2ID, trackKindAudio)
		}
		b.writeTrackTractor(ti)
	}
}

func (ti *trackInfo) hasUnmutedClips() bool {
	for _, e := range ti.entries {
		if e.c.SourcePath() != "" && e.c.HasAudio() {
			return true
		}
	}
	return false
}

type playlistMode int

const (
	allContent playlistMode = iota
)

func (b *builder) writeContentPlaylist(ti *trackInfo, playlistID string, _ playlistMode) {
	b.writeOpenTag("playlist", writeAttr("id", playlistID))
	if ti.kind == trackKindAudio {
		b.writeProperty("kdenlive:audio_track", "1")
	}

	cursor := time.Duration(0)
	for _, e := range ti.entries {
		if e.c.SourcePath() == "" {
			continue
		}

		// For audio tracks (BGM, narration), insert blanks during unmuted clip ranges
		// so the background music ducks under dialogue.
		startAt := e.startAt
		duration := e.c.Duration()

		if ti.kind == trackKindAudio {
			// Write the clip as segments split around unmuted ranges
			clipPos := startAt
			clipEnd := startAt + duration
			for _, ur := range b.unmutedRanges {
				if ur.start >= clipEnd || ur.end <= clipPos {
					continue
				}
				// There's overlap — write the part before the unmuted range, then a blank
				beforeDur := ur.start - clipPos
				if beforeDur > 0 {
					gap := clipPos - cursor
					if gap > 0 {
						b.writeEmptyTag("blank", writeAttr("length", formatTime(gap)))
					}
					chainID := b.chainIDForClip(e.c)
					trimStart := e.c.TrimStart() + (clipPos - startAt)
					b.writeEmptyTag("entry",
						writeAttr("in", formatTime(trimStart)),
						writeAttr("out", formatTime(trimStart+beforeDur)),
						writeAttr("producer", chainID),
					)
					cursor = clipPos + beforeDur
				}
				// Blank during the unmuted range
				unmutedDur := ur.end - ur.start
				if ur.start < clipPos {
					unmutedDur = ur.end - clipPos
				}
				if ur.end > clipEnd {
					unmutedDur = clipEnd - ur.start
				}
				if ur.start > clipPos {
					unmutedDur = ur.end - ur.start
				}
				// Recalculate unmutedDur precisely
				overlapStart := max(clipPos, ur.start)
				overlapEnd := min(clipEnd, ur.end)
				unmutedDur = overlapEnd - overlapStart
				if unmutedDur > 0 {
					b.writeEmptyTag("blank", writeAttr("length", formatTime(unmutedDur)))
					cursor = overlapEnd
				}
				clipPos = max(clipPos, ur.end)
			}
			// Write remaining part of the clip after all unmuted ranges
			if clipPos < clipEnd {
				remainingDur := clipEnd - clipPos
				gap := clipPos - cursor
				if gap > 0 {
					b.writeEmptyTag("blank", writeAttr("length", formatTime(gap)))
				}
				chainID := b.chainIDForClip(e.c)
				trimStart := e.c.TrimStart() + (clipPos - startAt)
				hasVolume := e.c.Volume() != 1.0
				if hasVolume {
					b.writeOpenTag("entry",
						writeAttr("in", formatTime(trimStart)),
						writeAttr("out", formatTime(trimStart+remainingDur)),
						writeAttr("producer", chainID),
					)
					fid := b.nextFilterID()
					b.writeOpenTag("filter",
						writeAttr("id", fid),
						writeAttr("in", formatTime(trimStart)),
						writeAttr("out", formatTime(trimStart+remainingDur)),
					)
					b.writeProperty("window", "75")
					b.writeProperty("max_gain", "20dB")
					b.writeProperty("mlt_service", "volume")
					b.writeProperty("gain", fmt.Sprintf("%g", e.c.Volume()))
					b.writeCloseTag("filter")
					b.writeCloseTag("entry")
				} else {
					b.writeEmptyTag("entry",
						writeAttr("in", formatTime(trimStart)),
						writeAttr("out", formatTime(trimStart+remainingDur)),
						writeAttr("producer", chainID),
					)
				}
				cursor = clipEnd
			}
			continue
		}

		gap := e.startAt - cursor
		if gap > 0 {
			b.writeEmptyTag("blank", writeAttr("length", formatTime(gap)))
		}

		chainID := b.chainIDForClip(e.c)
		trimStart := e.c.TrimStart()
		clipDur := e.c.Duration()
		entryIn := formatTime(trimStart)
		entryOut := formatTime(trimStart + clipDur)

		hasFilters := false
		if ti.kind == trackKindVideo {
			hasFilters = e.c.FadeInDuration() > 0 || e.c.FadeOutDuration() > 0
		}
		if e.c.Volume() != 1.0 {
			hasFilters = true
		}

		if hasFilters {
			b.writeOpenTag("entry",
				writeAttr("in", entryIn),
				writeAttr("out", entryOut),
				writeAttr("producer", chainID),
			)

			if ti.kind == trackKindVideo {
				if e.c.FadeInDuration() > 0 {
					fid := b.nextFilterID()
					b.writeOpenTag("filter",
						writeAttr("id", fid),
						writeAttr("in", formatTime(trimStart)),
						writeAttr("out", formatTime(trimStart+e.c.FadeInDuration())),
					)
					b.writeProperty("start", "1")
					b.writeProperty("level", "1")
					b.writeProperty("mlt_service", "brightness")
					b.writeProperty("kdenlive_id", "fade_from_black")
					b.writeProperty("alpha", "0=0;-1=1")
					b.writeCloseTag("filter")
				}

				if e.c.FadeOutDuration() > 0 {
					fid := b.nextFilterID()
					fadeOutStart := trimStart + clipDur - e.c.FadeOutDuration()
					b.writeOpenTag("filter",
						writeAttr("id", fid),
						writeAttr("in", formatTime(fadeOutStart)),
						writeAttr("out", formatTime(trimStart+clipDur)),
					)
					b.writeProperty("start", "1")
					b.writeProperty("level", "1")
					b.writeProperty("mlt_service", "brightness")
					b.writeProperty("kdenlive_id", "fade_to_black")
					b.writeProperty("alpha", "0=1;-1=0")
					b.writeCloseTag("filter")
				}
			}

			if e.c.Volume() != 1.0 {
				fid := b.nextFilterID()
				b.writeOpenTag("filter",
					writeAttr("id", fid),
					writeAttr("in", formatTime(trimStart)),
					writeAttr("out", formatTime(trimStart+clipDur)),
				)
				b.writeProperty("window", "75")
				b.writeProperty("max_gain", "20dB")
				b.writeProperty("mlt_service", "volume")
				b.writeProperty("gain", fmt.Sprintf("%g", e.c.Volume()))
				b.writeCloseTag("filter")
			}

			b.writeCloseTag("entry")
		} else {
			b.writeEmptyTag("entry",
				writeAttr("in", entryIn),
				writeAttr("out", entryOut),
				writeAttr("producer", chainID),
			)
		}

		cursor = e.startAt + clipDur
	}

	b.writeCloseTag("playlist")
}

func (b *builder) writeAudioCompanionPlaylist(ti *trackInfo) {
	b.writeOpenTag("playlist", writeAttr("id", ti.playlist1ID))
	b.writeProperty("kdenlive:audio_track", "1")

	// Walk the original video track entries. Audio clips become entries,
	// video-only clips become blanks of the same duration — this keeps everything in sync.
	cursor := time.Duration(0)
	for _, e := range ti.videoTrackEntries {
		if e.c.SourcePath() == "" {
			continue
		}

		duration := e.c.Duration()
		gap := e.startAt - cursor
		if gap > 0 {
			b.writeEmptyTag("blank", writeAttr("length", formatTime(gap)))
		}

		if e.c.HasAudio() {
			chainID := b.chainIDForClip(e.c)
			trimStart := e.c.TrimStart()
			entryIn := formatTime(trimStart)
			entryOut := formatTime(trimStart + duration)

			hasVolume := e.c.Volume() != 1.0
			if hasVolume {
				b.writeOpenTag("entry",
					writeAttr("in", entryIn),
					writeAttr("out", entryOut),
					writeAttr("producer", chainID),
				)
				fid := b.nextFilterID()
				b.writeOpenTag("filter",
					writeAttr("id", fid),
					writeAttr("in", formatTime(trimStart)),
					writeAttr("out", formatTime(trimStart+duration)),
				)
				b.writeProperty("window", "75")
				b.writeProperty("max_gain", "20dB")
				b.writeProperty("mlt_service", "volume")
				b.writeProperty("gain", fmt.Sprintf("%g", e.c.Volume()))
				b.writeCloseTag("filter")
				b.writeCloseTag("entry")
			} else {
				b.writeEmptyTag("entry",
					writeAttr("in", entryIn),
					writeAttr("out", entryOut),
					writeAttr("producer", chainID),
				)
			}
		} else {
			// Video-only clip: emit a blank of the same duration so audio clips stay in sync
			b.writeEmptyTag("blank", writeAttr("length", formatTime(duration)))
		}

		cursor = e.startAt + duration
	}

	b.writeCloseTag("playlist")
}

func (b *builder) writeEmptyPlaylist(id string, kind trackKind) {
	b.writeOpenTag("playlist", writeAttr("id", id))
	if kind == trackKindAudio {
		b.writeProperty("kdenlive:audio_track", "1")
	}
	b.writeCloseTag("playlist")
}

func (b *builder) writeTrackTractor(ti *trackInfo) {
	b.writeOpenTag("tractor", writeAttr("id", ti.tractorID))

	switch ti.kind {
	case trackKindVideo:
		b.writeProperty("kdenlive:trackheight", "67")
		b.writeProperty("kdenlive:timeline_active", "1")
		b.writeProperty("kdenlive:collapsed", "0")
		b.writeProperty("kdenlive:thumbs_format", "")
		b.writeProperty("kdenlive:audio_rec", "")

		b.writeEmptyTag("track",
			writeAttr("producer", ti.playlist1ID),
			writeAttr("hide", "audio"),
		)
		b.writeEmptyTag("track",
			writeAttr("producer", ti.playlist2ID),
			writeAttr("hide", "audio"),
		)

	case trackKindAudio:
		b.writeProperty("kdenlive:audio_track", "1")
		b.writeProperty("kdenlive:trackheight", "67")
		b.writeProperty("kdenlive:timeline_active", "1")
		b.writeProperty("kdenlive:collapsed", "0")
		b.writeProperty("kdenlive:thumbs_format", "")
		b.writeProperty("kdenlive:audio_rec", "")

		b.writeEmptyTag("track",
			writeAttr("producer", ti.playlist1ID),
			writeAttr("hide", "video"),
		)
		b.writeEmptyTag("track",
			writeAttr("producer", ti.playlist2ID),
			writeAttr("hide", "video"),
		)

		fid1 := b.nextFilterID()
		b.writeOpenTag("filter", writeAttr("id", fid1))
		b.writeProperty("window", "75")
		b.writeProperty("max_gain", "20dB")
		b.writeProperty("mlt_service", "volume")
		b.writeProperty("internal_added", "237")
		b.writeProperty("disable", "1")
		b.writeCloseTag("filter")

		fid2 := b.nextFilterID()
		b.writeOpenTag("filter", writeAttr("id", fid2))
		b.writeProperty("channel", "-1")
		b.writeProperty("mlt_service", "panner")
		b.writeProperty("internal_added", "237")
		b.writeProperty("start", "0.5")
		b.writeProperty("disable", "1")
		b.writeCloseTag("filter")

		fid3 := b.nextFilterID()
		b.writeOpenTag("filter", writeAttr("id", fid3))
		b.writeProperty("iec_scale", "0")
		b.writeProperty("mlt_service", "audiolevel")
		b.writeProperty("dbpeak", "1")
		b.writeProperty("disable", "1")
		b.writeCloseTag("filter")

	case trackKindAudioCompanion:
		b.writeProperty("kdenlive:audio_track", "1")
		b.writeProperty("kdenlive:trackheight", "67")
		b.writeProperty("kdenlive:timeline_active", "1")
		b.writeProperty("kdenlive:collapsed", "0")
		b.writeProperty("kdenlive:thumbs_format", "")
		b.writeProperty("kdenlive:audio_rec", "")

		b.writeEmptyTag("track",
			writeAttr("producer", ti.playlist1ID),
			writeAttr("hide", "video"),
		)
		b.writeEmptyTag("track",
			writeAttr("producer", ti.playlist2ID),
			writeAttr("hide", "video"),
		)

		fid1 := b.nextFilterID()
		b.writeOpenTag("filter", writeAttr("id", fid1))
		b.writeProperty("window", "75")
		b.writeProperty("max_gain", "20dB")
		b.writeProperty("mlt_service", "volume")
		b.writeProperty("internal_added", "237")
		b.writeProperty("disable", "1")
		b.writeCloseTag("filter")

		fid2 := b.nextFilterID()
		b.writeOpenTag("filter", writeAttr("id", fid2))
		b.writeProperty("channel", "-1")
		b.writeProperty("mlt_service", "panner")
		b.writeProperty("internal_added", "237")
		b.writeProperty("start", "0.5")
		b.writeProperty("disable", "1")
		b.writeCloseTag("filter")

		fid3 := b.nextFilterID()
		b.writeOpenTag("filter", writeAttr("id", fid3))
		b.writeProperty("iec_scale", "0")
		b.writeProperty("mlt_service", "audiolevel")
		b.writeProperty("dbpeak", "1")
		b.writeProperty("disable", "1")
		b.writeCloseTag("filter")
	}

	b.writeCloseTag("tractor")
}

func (b *builder) writeTimelineTractor(timelineDur time.Duration) {
	hasVideo := false
	hasAudio := false
	for _, ti := range b.tracks {
		if ti.kind == trackKindVideo {
			hasVideo = true
		} else {
			hasAudio = true
		}
	}

	b.writeOpenTag("tractor",
		writeAttr("id", b.uuid),
		writeAttr("in", "00:00:00.000"),
		writeAttr("out", formatTime(timelineDur)),
	)

	b.writeProperty("kdenlive:uuid", b.uuid)
	b.writeProperty("kdenlive:clipname", "Sequence 1")
	b.writeProperty("kdenlive:sequenceproperties.hasAudio", boolStr(hasAudio))
	b.writeProperty("kdenlive:sequenceproperties.hasVideo", boolStr(hasVideo))
	b.writeProperty("kdenlive:sequenceproperties.activeTrack", "2")
	b.writeProperty("kdenlive:sequenceproperties.tracksCount", fmt.Sprintf("%d", len(b.tracks)))
	b.writeProperty("kdenlive:sequenceproperties.documentuuid", b.uuid)
	b.writeProperty("kdenlive:duration", "00:00:00;01")
	b.writeProperty("kdenlive:maxduration", "1")
	b.writeProperty("kdenlive:producer_type", "17")
	b.writeProperty("kdenlive:id", "3")
	b.writeProperty("kdenlive:clip_type", "0")
	b.writeProperty("kdenlive:file_hash", "ceb20492568cd0ec56711e5d15117ef3")
	b.writeProperty("kdenlive:folderid", "2")
	b.writeProperty("kdenlive:markers", "[\n]\n")
	b.writeProperty("kdenlive:sequenceproperties.audioTarget", "1")
	b.writeProperty("kdenlive:sequenceproperties.disablepreview", "0")
	b.writeProperty("kdenlive:sequenceproperties.position", "0")
	b.writeProperty("kdenlive:sequenceproperties.scrollPos", "0")
	b.writeProperty("kdenlive:sequenceproperties.tracks", fmt.Sprintf("%d", len(b.tracks)))
	b.writeProperty("kdenlive:sequenceproperties.verticalzoom", "1")
	if hasVideo {
		videoTrackIdx := -1
		for i, ti := range b.tracks {
			if ti.kind == trackKindVideo {
				videoTrackIdx = i
				break
			}
		}
		if videoTrackIdx >= 0 {
			b.writeProperty("kdenlive:sequenceproperties.videoTarget", fmt.Sprintf("%d", videoTrackIdx+1))
		}
	}
	b.writeProperty("kdenlive:sequenceproperties.zonein", "0")
	b.writeProperty("kdenlive:sequenceproperties.zoneout", "75")
	b.writeProperty("kdenlive:sequenceproperties.zoom", "8")
	b.writeProperty("kdenlive:sequenceproperties.groups", "[\n]\n")
	b.writeProperty("kdenlive:sequenceproperties.guides", "[\n]\n")

	b.writeEmptyTag("track", writeAttr("producer", "producer0"))

	for i, ti := range b.tracks {
		tid := b.nextTransID()
		b.writeOpenTag("transition", writeAttr("id", tid))
		b.writeProperty("a_track", "0")
		b.writeProperty("b_track", fmt.Sprintf("%d", i+1))

		if ti.kind == trackKindAudio || ti.kind == trackKindAudioCompanion {
			b.writeProperty("mlt_service", "mix")
			b.writeProperty("kdenlive_id", "mix")
			b.writeProperty("internal_added", "237")
			b.writeProperty("always_active", "1")
			b.writeProperty("accepts_blanks", "1")
			b.writeProperty("sum", "1")
		} else {
			b.writeProperty("compositing", "0")
			b.writeProperty("distort", "0")
			b.writeProperty("rotate_center", "0")
			b.writeProperty("mlt_service", "qtblend")
			b.writeProperty("kdenlive_id", "qtblend")
			b.writeProperty("internal_added", "237")
			b.writeProperty("always_active", "1")
		}
		b.writeCloseTag("transition")
	}

	fid1 := b.nextFilterID()
	b.writeOpenTag("filter", writeAttr("id", fid1))
	b.writeProperty("window", "75")
	b.writeProperty("max_gain", "20dB")
	b.writeProperty("mlt_service", "volume")
	b.writeProperty("internal_added", "237")
	b.writeProperty("disable", "1")
	b.writeCloseTag("filter")

	fid2 := b.nextFilterID()
	b.writeOpenTag("filter", writeAttr("id", fid2))
	b.writeProperty("channel", "-1")
	b.writeProperty("mlt_service", "panner")
	b.writeProperty("internal_added", "237")
	b.writeProperty("start", "0.5")
	b.writeProperty("disable", "1")
	b.writeCloseTag("filter")

	for _, ti := range b.tracks {
		b.writeEmptyTag("track", writeAttr("producer", ti.tractorID))
	}

	b.writeCloseTag("tractor")
}

func (b *builder) writeMainBin() {
	b.writeOpenTag("playlist", writeAttr("id", "main_bin"))

	b.writeProperty("kdenlive:folder.-1.2", "Sequences")
	b.writeProperty("kdenlive:sequenceFolder", "2")
	b.writeProperty("kdenlive:docproperties.audioChannels", "2")
	b.writeProperty("kdenlive:docproperties.binsort", "0")
	b.writeProperty("kdenlive:docproperties.documentid", b.docID)
	b.writeProperty("kdenlive:docproperties.enableTimelineZone", "0")
	b.writeProperty("kdenlive:docproperties.enableexternalproxy", "0")
	b.writeProperty("kdenlive:docproperties.enableproxy", "0")
	b.writeProperty("kdenlive:docproperties.externalproxyparams", "")
	b.writeProperty("kdenlive:docproperties.generateimageproxy", "0")
	b.writeProperty("kdenlive:docproperties.generateproxy", "0")
	b.writeProperty("kdenlive:docproperties.guidesCategories", guidesJSON)
	b.writeProperty("kdenlive:docproperties.kdenliveversion", "24.05.1")
	b.writeProperty("kdenlive:docproperties.previewextension", "")
	b.writeProperty("kdenlive:docproperties.previewparameters", "")
	b.writeProperty("kdenlive:docproperties.profile", profileName(b.tl.Config()))
	b.writeProperty("kdenlive:docproperties.proxyextension", "")
	b.writeProperty("kdenlive:docproperties.proxyimageminsize", "2000")
	b.writeProperty("kdenlive:docproperties.proxyimagesize", "800")
	b.writeProperty("kdenlive:docproperties.proxyminsize", "1000")
	b.writeProperty("kdenlive:docproperties.proxyparams", "")
	b.writeProperty("kdenlive:docproperties.proxyresize", "640")
	b.writeProperty("kdenlive:docproperties.seekOffset", "18000")
	b.writeProperty("kdenlive:docproperties.uuid", b.uuid)
	b.writeProperty("kdenlive:docproperties.version", "1.1")
	b.writeProperty("kdenlive:expandedFolders", "")
	b.writeProperty("kdenlive:binZoom", "4")
	b.writeProperty("kdenlive:extraBins", "project_bin:-1:0")
	b.writeProperty("kdenlive:documentnotes", "")
	b.writeProperty("kdenlive:docproperties.opensequences", b.uuid)
	b.writeProperty("kdenlive:docproperties.activetimeline", b.uuid)
	b.writeProperty("xml_retain", "1")

	b.writeEmptyTag("entry",
		writeAttr("in", "00:00:00.000"),
		writeAttr("out", "00:00:00.000"),
		writeAttr("producer", b.uuid),
	)

	for _, ci := range b.chainList {
		b.writeEmptyTag("entry",
			writeAttr("in", "00:00:00.000"),
			writeAttr("out", "00:00:00.000"),
			writeAttr("producer", ci.id),
		)
	}

	b.writeCloseTag("playlist")
}

func (b *builder) writeFinalTractor() {
	b.writeOpenTag("tractor",
		writeAttr("id", "final_tractor"),
		writeAttr("in", "00:00:00.000"),
		writeAttr("out", "00:00:00.000"),
	)
	b.writeProperty("kdenlive:projectTractor", "1")
	b.writeEmptyTag("track",
		writeAttr("in", "00:00:00.000"),
		writeAttr("out", "00:00:00.000"),
		writeAttr("producer", b.uuid),
	)
	b.writeCloseTag("tractor")
}

func formatTime(d time.Duration) string {
	totalSec := d.Seconds()
	if totalSec < 0 {
		totalSec = 0
	}
	hours := int(totalSec) / 3600
	remaining := math.Mod(totalSec, 3600)
	minutes := int(remaining) / 60
	remaining = math.Mod(remaining, 60)
	seconds := int(remaining)
	millis := int(math.Round((remaining-float64(seconds))*1000)) % 1000
	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, millis)
}

func boolStr(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func profileName(cfg timeline.Config) string {
	switch {
	case cfg.Width == 1920 && cfg.Height == 1080 && cfg.FPS == 29.97:
		return "atsc_1080p_2997"
	case cfg.Width == 1920 && cfg.Height == 1080 && cfg.FPS == 30:
		return "atsc_1080p_30"
	case cfg.Width == 1920 && cfg.Height == 1080 && cfg.FPS == 25:
		return "atsc_1080p_25"
	case cfg.Width == 1920 && cfg.Height == 1080 && cfg.FPS == 24:
		return "atsc_1080p_24"
	case cfg.Width == 3840 && cfg.Height == 2160 && cfg.FPS == 30:
		return "atsc_2160p_30"
	default:
		return fmt.Sprintf("%dx%d_%gfps", cfg.Width, cfg.Height, cfg.FPS)
	}
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func escapeAttr(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func escapeText(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func joinAttrs(attrs []string) string {
	if len(attrs) == 0 {
		return ""
	}
	return " " + strings.Join(attrs, " ")
}

func joinAttrs2(tag string, attrs []string) string {
	if len(attrs) == 0 {
		return tag
	}
	return tag + " " + strings.Join(attrs, " ")
}

func generateUUID() string {
	var buf [16]byte
	rand.Read(buf[:])
	return fmt.Sprintf("{%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x}",
		buf[0], buf[1], buf[2], buf[3],
		buf[4], buf[5],
		buf[6], buf[7],
		buf[8], buf[9],
		buf[10], buf[11], buf[12], buf[13], buf[14], buf[15],
	)
}

var guidesJSON = "[\n" +
	"    {\n" +
	"        \"color\": \"#9b59b6\",\n" +
	"        \"comment\": \"Category 1\",\n" +
	"        \"index\": 0\n" +
	"    },\n" +
	"    {\n" +
	"        \"color\": \"#3daee9\",\n" +
	"        \"comment\": \"Category 2\",\n" +
	"        \"index\": 1\n" +
	"    },\n" +
	"    {\n" +
	"        \"color\": \"#1abc9c\",\n" +
	"        \"comment\": \"Category 3\",\n" +
	"        \"index\": 2\n" +
	"    },\n" +
	"    {\n" +
	"        \"color\": \"#1cdc9a\",\n" +
	"        \"comment\": \"Category 4\",\n" +
	"        \"index\": 3\n" +
	"    },\n" +
	"    {\n" +
	"        \"color\": \"#c9ce3b\",\n" +
	"        \"comment\": \"Category 5\",\n" +
	"        \"index\": 4\n" +
	"    },\n" +
	"    {\n" +
	"        \"color\": \"#fdbc4b\",\n" +
	"        \"comment\": \"Category 6\",\n" +
	"        \"index\": 5\n" +
	"    },\n" +
	"    {\n" +
	"        \"color\": \"#f39c1f\",\n" +
	"        \"comment\": \"Category 7\",\n" +
	"        \"index\": 6\n" +
	"    },\n" +
	"    {\n" +
	"        \"color\": \"#f47750\",\n" +
	"        \"comment\": \"Category 8\",\n" +
	"        \"index\": 7\n" +
	"    },\n" +
	"    {\n" +
	"        \"color\": \"#da4453\",\n" +
	"        \"comment\": \"Category 9\",\n" +
	"        \"index\": 8\n" +
	"    }\n" +
	"]"