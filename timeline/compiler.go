package timeline

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ahmedhodiani/gomontage/clip"
	"github.com/ahmedhodiani/gomontage/effects"
	"github.com/ahmedhodiani/gomontage/engine"
)

// Compiler transforms a Timeline into an FFmpeg engine.Graph that can be
// built into a command and executed.
type Compiler struct {
	timeline *Timeline
	graph    *engine.Graph

	// inputMap tracks which source files have already been added as inputs,
	// mapping source path -> input node. This avoids duplicate -i entries.
	inputMap map[string]*engine.Node

	// labelMap tracks which node produces each named label. This allows
	// findFilterByLabel to look up nodes by their output label without
	// requiring the label to exist as an actual graph edge yet.
	labelMap map[string]*engine.Node

	// labelCounter generates unique filter graph labels.
	labelCounter int
}

// NewCompiler creates a compiler for the given timeline.
func NewCompiler(tl *Timeline) *Compiler {
	return &Compiler{
		timeline: tl,
		graph:    engine.NewGraph(),
		inputMap: make(map[string]*engine.Node),
		labelMap: make(map[string]*engine.Node),
	}
}

// nextLabel generates a unique label for the filter graph.
func (c *Compiler) nextLabel(prefix string) string {
	c.labelCounter++
	return fmt.Sprintf("%s%d", prefix, c.labelCounter)
}

// nextLabelFor generates a unique label and registers the given node as its producer.
func (c *Compiler) nextLabelFor(prefix string, node *engine.Node) string {
	label := c.nextLabel(prefix)
	c.labelMap[label] = node
	return label
}

// getOrAddInput returns the input node for a source path, adding it if needed.
func (c *Compiler) getOrAddInput(path string) *engine.Node {
	if node, ok := c.inputMap[path]; ok {
		return node
	}
	node := c.graph.AddInput(path)
	c.inputMap[path] = node
	return node
}

// trackStream holds the compiled video and audio output labels for a single
// video track after its clips have been processed and concatenated.
type trackStream struct {
	video string // final video label for this track
	audio string // final audio label for this track (may be empty)
}

// Compile transforms the timeline into an FFmpeg filter graph.
// The resulting graph can be built into a Command via engine.BuildCommand.
//
// Video tracks are composited in layer order: track 0 is the bottom (background)
// and each subsequent track is overlaid on top using FFmpeg's overlay filter.
// Transparent PNGs and RGBA video sources are supported via overlay=format=auto.
//
// All audio tracks (including audio embedded in video clips) are mixed together
// using amix, so narration, background music, and sound effects all stack.
func (c *Compiler) Compile(outputPath string, outputParams map[string]string) (*engine.Graph, error) {
	if err := c.timeline.Validate(); err != nil {
		return nil, fmt.Errorf("timeline validation failed: %w", err)
	}

	cfg := c.timeline.Config()
	timelineDur := c.timeline.Duration()

	// Phase 1: Compile each video track independently into its own stream.
	// Each track's clips are sorted, gap-filled, and concatenated in isolation.
	// This produces one video label (and optionally one audio label) per track.
	var trackStreams []trackStream
	for _, track := range c.timeline.videoTracks {
		if len(track.Entries()) == 0 {
			continue
		}
		ts, err := c.compileTrack(track, cfg, timelineDur)
		if err != nil {
			return nil, fmt.Errorf("track %q: %w", track.Name(), err)
		}
		trackStreams = append(trackStreams, ts)
	}

	// Phase 2: Composite video tracks together.
	// Single track → use directly (no overlay needed, keeps old behaviour).
	// Multiple tracks → chain overlay filters: track[0] is base, each
	// subsequent track is overlaid on top with format=auto (supports RGBA/alpha).
	finalVideoLabel := ""
	var videoAudioLabels []string // audio from video tracks, to be mixed later

	switch len(trackStreams) {
	case 0:
		// No video tracks — nothing to composite.
	case 1:
		finalVideoLabel = trackStreams[0].video
		if trackStreams[0].audio != "" {
			videoAudioLabels = append(videoAudioLabels, trackStreams[0].audio)
		}
	default:
		// Collect audio from all tracks.
		for _, ts := range trackStreams {
			if ts.audio != "" {
				videoAudioLabels = append(videoAudioLabels, ts.audio)
			}
		}
		// Chain overlay filters bottom-up.
		finalVideoLabel = c.overlayTracks(trackStreams)
	}

	// Phase 3: Process independent audio tracks (narration, bgm, sfx, etc.).
	// All clips within each audio track are processed independently and their
	// labels collected; they will all be mixed together in Phase 4.
	var audioLabels []string
	// Prepend audio from video tracks so it is mixed with independent audio.
	audioLabels = append(audioLabels, videoAudioLabels...)
	for _, track := range c.timeline.audioTracks {
		for _, entry := range track.Entries() {
			label, err := c.compileAudioEntry(entry)
			if err != nil {
				return nil, fmt.Errorf("audio track %q: %w", track.Name(), err)
			}
			if label != "" {
				audioLabels = append(audioLabels, label)
			}
		}
	}

	// Phase 4: Mix all audio streams together.
	// amix naturally stacks any number of streams, so narration + bgm + sfx
	// all combine correctly regardless of how many audio tracks there are.
	finalAudioLabel := ""
	if len(audioLabels) == 1 {
		finalAudioLabel = audioLabels[0]
	} else if len(audioLabels) > 1 {
		finalAudioLabel = c.mixAudio(audioLabels)
	}

	// Phase 5: Create output node and wire up final streams.
	if outputParams == nil {
		outputParams = make(map[string]string)
	}
	output := c.graph.AddOutput(outputPath, outputParams)

	if finalVideoLabel != "" {
		c.graph.Connect(
			c.findFilterByLabel(finalVideoLabel),
			output,
			finalVideoLabel,
			engine.StreamVideo,
		)
	}
	if finalAudioLabel != "" {
		c.graph.Connect(
			c.findFilterByLabel(finalAudioLabel),
			output,
			finalAudioLabel,
			engine.StreamAudio,
		)
	}

	return c.graph, nil
}

// compileTrack compiles a single VideoTrack into a trackStream.
// It processes each clip, sorts by StartAt, inserts black gap clips for any
// time gaps (including before the first clip), and concatenates everything into
// a single video stream. Audio embedded in video clips is concatenated in sync.
//
// timelineDur is the total duration of the timeline. If this track is shorter,
// it is padded with black (and silence) to match, so that overlay compositing
// works correctly for the full duration.
func (c *Compiler) compileTrack(track *VideoTrack, cfg Config, timelineDur time.Duration) (trackStream, error) {
	// Step 1: compile every clip entry on this track.
	var labels []clipLabel
	for _, entry := range track.Entries() {
		cls, err := c.compileVideoEntry(entry, cfg)
		if err != nil {
			return trackStream{}, err
		}
		// Compile embedded audio from video clips.
		for i, cl := range cls {
			if cl.entry.Clip.HasAudio() && cl.entry.Clip.SourcePath() != "" {
				audioLabel, err := c.compileAudioFromVideoEntry(cl.entry)
				if err != nil {
					return trackStream{}, fmt.Errorf("audio: %w", err)
				}
				cls[i].audio = audioLabel
			}
		}
		labels = append(labels, cls...)
	}

	// Step 2: sort by StartAt and insert gap clips for any time gaps.
	if len(labels) > 1 {
		sort.SliceStable(labels, func(i, j int) bool {
			return labels[i].entry.StartAt < labels[j].entry.StartAt
		})
	}
	labels = c.insertVideoGaps(labels, cfg)

	// Step 3: concatenate (or pass through if single clip).
	var videoLabel, audioLabel string
	switch len(labels) {
	case 0:
		return trackStream{}, nil
	case 1:
		videoLabel = labels[0].video
		audioLabel = labels[0].audio
		// Single-clip: if it starts at a non-zero time, pad the front.
		firstStartAt := labels[0].entry.StartAt
		if firstStartAt > 0 {
			tpadNode := c.graph.AddFilter("tpad", map[string]string{
				"start_duration": formatSeconds(firstStartAt),
				"color":          "black",
			})
			tpadLabel := c.nextLabelFor("vpad", tpadNode)
			c.graph.Connect(c.findFilterByLabel(videoLabel), tpadNode, videoLabel, engine.StreamVideo)
			videoLabel = tpadLabel

			if audioLabel != "" {
				delayNode := c.graph.AddFilter("adelay", map[string]string{
					"delays": fmt.Sprintf("%d|%d", firstStartAt.Milliseconds(), firstStartAt.Milliseconds()),
				})
				delayLabel := c.nextLabelFor("vadel", delayNode)
				c.graph.Connect(c.findFilterByLabel(audioLabel), delayNode, audioLabel, engine.StreamAudio)
				audioLabel = delayLabel
			}
		}
	default:
		videoLabel, audioLabel = c.concatVideoClips(labels)
	}

	// Step 4: if the track is shorter than the full timeline, pad the tail with
	// black frames (and silence) so overlay compositing covers the whole duration.
	trackEnd := track.End()
	if timelineDur > trackEnd {
		tailDur := timelineDur - trackEnd
		tpadNode := c.graph.AddFilter("tpad", map[string]string{
			"stop_duration": formatSeconds(tailDur),
			"color":         "black",
		})
		tpadLabel := c.nextLabelFor("vtail", tpadNode)
		c.graph.Connect(c.findFilterByLabel(videoLabel), tpadNode, videoLabel, engine.StreamVideo)
		videoLabel = tpadLabel

		if audioLabel != "" {
			apadNode := c.graph.AddFilter("apad", map[string]string{
				"pad_dur": formatSeconds(tailDur),
			})
			apadLabel := c.nextLabelFor("atail", apadNode)
			c.graph.Connect(c.findFilterByLabel(audioLabel), apadNode, audioLabel, engine.StreamAudio)
			audioLabel = apadLabel
		}
	}

	return trackStream{video: videoLabel, audio: audioLabel}, nil
}

// overlayTracks chains overlay filters across all track streams.
// trackStreams[0] is the bottom (background) layer; each subsequent stream is
// composited on top. overlay=format=auto is used so that RGBA sources (e.g.
// transparent PNGs) are handled correctly — the alpha channel is respected.
func (c *Compiler) overlayTracks(trackStreams []trackStream) string {
	// Start with the bottom track as the base.
	baseLabel := trackStreams[0].video

	for i := 1; i < len(trackStreams); i++ {
		topLabel := trackStreams[i].video

		overlayNode := c.graph.AddFilter("overlay", map[string]string{
			"x":      "0",
			"y":      "0",
			"format": "auto",
		})
		overlayLabel := c.nextLabelFor("voverlay", overlayNode)

		// overlay expects [base][top] as its two inputs.
		c.graph.Connect(c.findFilterByLabel(baseLabel), overlayNode, baseLabel, engine.StreamVideo)
		c.graph.Connect(c.findFilterByLabel(topLabel), overlayNode, topLabel, engine.StreamVideo)

		baseLabel = overlayLabel
	}

	return baseLabel
}

// clipLabel holds the output labels for a processed clip.
type clipLabel struct {
	video string
	audio string
	entry Placement
}

// compileVideoEntry processes a single video clip placement into filter nodes.
func (c *Compiler) compileVideoEntry(entry Placement, cfg Config) ([]clipLabel, error) {
	cl := entry.Clip
	path := cl.SourcePath()

	var currentLabel string
	var lastNode *engine.Node

	if path == "" {
		// Generated clips (color, text) — use source filters.
		genLabel, genNode, err := c.compileGeneratedClip(entry, cfg)
		if err != nil {
			return nil, err
		}
		currentLabel = genLabel
		lastNode = genNode
	} else {
		inputNode := c.getOrAddInput(path)
		inputIdx := c.graph.InputIndex(inputNode)

		// Image clips need special input-level flags to produce a video stream
		// from a single still frame. Without these, FFmpeg reads one frame and
		// the concat filter gets a near-zero-duration segment.
		if cl.ClipType() == clip.TypeImage {
			inputNode.Params["-loop"] = "1"
			inputNode.Params["-framerate"] = fmt.Sprintf("%g", cfg.FPS)
			inputNode.Params["-t"] = formatSeconds(cl.Duration())
		}

		// Start with the raw video stream.
		currentLabel = fmt.Sprintf("%d:v", inputIdx)
		lastNode = inputNode

		// Apply trim if needed.
		if cl.IsTrimmed() {
			trimNode := c.graph.AddFilter("trim", map[string]string{
				"start": formatSeconds(cl.TrimStart()),
				"end":   formatSeconds(cl.TrimEnd()),
			})
			trimLabel := c.nextLabelFor("vtrim", trimNode)
			c.graph.Connect(lastNode, trimNode, currentLabel, engine.StreamVideo)

			// After trim, reset PTS to start from 0.
			setptsNode := c.graph.AddFilter("setpts", map[string]string{
				"expr": "PTS-STARTPTS",
			})
			setptsLabel := c.nextLabelFor("vpts", setptsNode)
			c.graph.Connect(trimNode, setptsNode, trimLabel, engine.StreamVideo)

			currentLabel = setptsLabel
			lastNode = setptsNode
		}

		// Scale file-based clips to the timeline resolution so all segments
		// match when concatenated. The timeline config is the single source
		// of truth for output dimensions, just like a real NLE.
		if cfg.Width > 0 && cfg.Height > 0 {
			scaleNode := c.graph.AddFilter("scale", map[string]string{
				"w": fmt.Sprintf("%d", cfg.Width),
				"h": fmt.Sprintf("%d", cfg.Height),
			})
			scaleLabel := c.nextLabelFor("vscale", scaleNode)
			c.graph.Connect(lastNode, scaleNode, currentLabel, engine.StreamVideo)
			currentLabel = scaleLabel
			lastNode = scaleNode
		}

		// Image clips need pixel format normalization to yuv420p so they can
		// be concatenated with video clips (which are typically yuv420p).
		// Generated clips (color, text) already get this in compileGeneratedClip.
		if cl.ClipType() == clip.TypeImage {
			fmtNode := c.graph.AddFilter("format", map[string]string{
				"pix_fmts": "yuv420p",
			})
			fmtLabel := c.nextLabelFor("imgfmt", fmtNode)
			c.graph.Connect(lastNode, fmtNode, currentLabel, engine.StreamVideo)
			currentLabel = fmtLabel
			lastNode = fmtNode
		}
	}

	// Apply fade in (works for both generated and file-based clips).
	if cl.FadeInDuration() > 0 {
		fadeNode := c.graph.AddFilter("fade", map[string]string{
			"t":  "in",
			"st": "0",
			"d":  formatSeconds(cl.FadeInDuration()),
		})
		fadeLabel := c.nextLabelFor("vfin", fadeNode)
		c.graph.Connect(lastNode, fadeNode, currentLabel, engine.StreamVideo)
		currentLabel = fadeLabel
		lastNode = fadeNode
	}

	// Apply fade out (works for both generated and file-based clips).
	if cl.FadeOutDuration() > 0 {
		fadeOutStart := cl.Duration() - cl.FadeOutDuration()
		fadeNode := c.graph.AddFilter("fade", map[string]string{
			"t":  "out",
			"st": formatSeconds(fadeOutStart),
			"d":  formatSeconds(cl.FadeOutDuration()),
		})
		fadeLabel := c.nextLabelFor("vfout", fadeNode)
		c.graph.Connect(lastNode, fadeNode, currentLabel, engine.StreamVideo)
		currentLabel = fadeLabel
		lastNode = fadeNode
	}

	// Apply composable video effects from the clip's Effects() list.
	for _, e := range cl.Effects() {
		if e.Target() != effects.TargetVideo && e.Target() != effects.TargetBoth {
			continue
		}
		filterNode := c.graph.AddFilter(e.FilterName(), e.FilterParams())
		filterLabel := c.nextLabelFor("vfx", filterNode)
		c.graph.Connect(lastNode, filterNode, currentLabel, engine.StreamVideo)
		currentLabel = filterLabel
		lastNode = filterNode
	}

	return []clipLabel{{
		video: currentLabel,
		entry: entry,
	}}, nil
}

// compileGeneratedClip handles color and text clips that don't come from files.
// It returns the output label and the last node in the filter chain, so that
// the caller can continue applying effects (fades, etc.) to the generated clip.
func (c *Compiler) compileGeneratedClip(entry Placement, cfg Config) (string, *engine.Node, error) {
	cl := entry.Clip

	switch v := cl.(type) {
	case *clip.ColorClip:
		colorHex := strings.TrimPrefix(v.Color, "#")
		w := v.Width()
		h := v.Height()
		if w == 0 {
			w = cfg.Width
		}
		if h == 0 {
			h = cfg.Height
		}
		node := c.graph.AddFilter("color", map[string]string{
			"c": colorHex,
			"s": fmt.Sprintf("%dx%d", w, h),
			"d": formatSeconds(v.Duration()),
			"r": fmt.Sprintf("%g", cfg.FPS),
		})
		label := c.nextLabelFor("gen", node)

		// Normalize pixel format to yuv420p so generated clips can be
		// concatenated with real video (which is typically yuv420p).
		fmtNode := c.graph.AddFilter("format", map[string]string{
			"pix_fmts": "yuv420p",
		})
		fmtLabel := c.nextLabelFor("genfmt", fmtNode)
		c.graph.Connect(node, fmtNode, label, engine.StreamVideo)

		return fmtLabel, fmtNode, nil

	case *clip.TextClip:
		// Text clips use the drawtext filter on a solid color background.
		bgNode := c.graph.AddFilter("color", map[string]string{
			"c": "black",
			"s": fmt.Sprintf("%dx%d", cfg.Width, cfg.Height),
			"d": formatSeconds(v.Duration()),
			"r": fmt.Sprintf("%g", cfg.FPS),
		})
		bgLabel := c.nextLabelFor("txtbg", bgNode)

		textNode := c.graph.AddFilter("drawtext", map[string]string{
			"text":      v.Text,
			"fontsize":  fmt.Sprintf("%d", v.Style.Size),
			"fontcolor": v.Style.Color,
		})
		if v.Style.Font != "" {
			textNode.Params["fontfile"] = v.Style.Font
		}
		textLabel := c.nextLabelFor("gen", textNode)
		c.graph.Connect(bgNode, textNode, bgLabel, engine.StreamVideo)

		// Normalize pixel format to yuv420p for concat compatibility.
		fmtNode := c.graph.AddFilter("format", map[string]string{
			"pix_fmts": "yuv420p",
		})
		fmtLabel := c.nextLabelFor("genfmt", fmtNode)
		c.graph.Connect(textNode, fmtNode, textLabel, engine.StreamVideo)

		return fmtLabel, fmtNode, nil

	default:
		return "", nil, fmt.Errorf("unsupported generated clip type: %T", cl)
	}
}

// compileAudioEntry processes a single audio clip placement into filter nodes.
func (c *Compiler) compileAudioEntry(entry Placement) (string, error) {
	cl := entry.Clip
	path := cl.SourcePath()

	if path == "" {
		return "", nil // No source for generated clips.
	}

	inputNode := c.getOrAddInput(path)
	inputIdx := c.graph.InputIndex(inputNode)

	currentLabel := fmt.Sprintf("%d:a", inputIdx)
	var lastNode *engine.Node = inputNode

	// Apply trim.
	if cl.IsTrimmed() {
		trimNode := c.graph.AddFilter("atrim", map[string]string{
			"start": formatSeconds(cl.TrimStart()),
			"end":   formatSeconds(cl.TrimEnd()),
		})
		trimLabel := c.nextLabelFor("atrim", trimNode)
		c.graph.Connect(lastNode, trimNode, currentLabel, engine.StreamAudio)

		aptsNode := c.graph.AddFilter("asetpts", map[string]string{
			"expr": "PTS-STARTPTS",
		})
		aptsLabel := c.nextLabelFor("apts", aptsNode)
		c.graph.Connect(trimNode, aptsNode, trimLabel, engine.StreamAudio)

		currentLabel = aptsLabel
		lastNode = aptsNode
	}

	// Apply volume.
	if cl.Volume() != 1.0 {
		volNode := c.graph.AddFilter("volume", map[string]string{
			"volume": fmt.Sprintf("%g", cl.Volume()),
		})
		volLabel := c.nextLabelFor("avol", volNode)
		c.graph.Connect(lastNode, volNode, currentLabel, engine.StreamAudio)
		currentLabel = volLabel
		lastNode = volNode
	}

	// Apply audio fade in.
	if cl.FadeInDuration() > 0 {
		fadeNode := c.graph.AddFilter("afade", map[string]string{
			"t":  "in",
			"st": "0",
			"d":  formatSeconds(cl.FadeInDuration()),
		})
		fadeLabel := c.nextLabelFor("afin", fadeNode)
		c.graph.Connect(lastNode, fadeNode, currentLabel, engine.StreamAudio)
		currentLabel = fadeLabel
		lastNode = fadeNode
	}

	// Apply audio fade out.
	if cl.FadeOutDuration() > 0 {
		fadeOutStart := cl.Duration() - cl.FadeOutDuration()
		fadeNode := c.graph.AddFilter("afade", map[string]string{
			"t":  "out",
			"st": formatSeconds(fadeOutStart),
			"d":  formatSeconds(cl.FadeOutDuration()),
		})
		fadeLabel := c.nextLabelFor("afout", fadeNode)
		c.graph.Connect(lastNode, fadeNode, currentLabel, engine.StreamAudio)
		currentLabel = fadeLabel
		lastNode = fadeNode
	}

	// Apply composable audio effects from the clip's Effects() list.
	for _, e := range cl.Effects() {
		if e.Target() != effects.TargetAudio && e.Target() != effects.TargetBoth {
			continue
		}
		currentLabel, lastNode = c.compileAudioEffect(e, currentLabel, lastNode, "afx")
	}

	// Apply delay if the clip doesn't start at 0.
	if entry.StartAt > 0 {
		delayNode := c.graph.AddFilter("adelay", map[string]string{
			"delays": fmt.Sprintf("%d|%d", entry.StartAt.Milliseconds(), entry.StartAt.Milliseconds()),
		})
		delayLabel := c.nextLabelFor("adel", delayNode)
		c.graph.Connect(lastNode, delayNode, currentLabel, engine.StreamAudio)
		currentLabel = delayLabel
	}

	return currentLabel, nil
}

// compileAudioFromVideoEntry extracts and processes the audio from a video clip entry.
func (c *Compiler) compileAudioFromVideoEntry(entry Placement) (string, error) {
	cl := entry.Clip
	path := cl.SourcePath()

	inputNode := c.getOrAddInput(path)
	inputIdx := c.graph.InputIndex(inputNode)

	currentLabel := fmt.Sprintf("%d:a", inputIdx)
	var lastNode *engine.Node = inputNode

	// Apply trim.
	if cl.IsTrimmed() {
		trimNode := c.graph.AddFilter("atrim", map[string]string{
			"start": formatSeconds(cl.TrimStart()),
			"end":   formatSeconds(cl.TrimEnd()),
		})
		trimLabel := c.nextLabelFor("vatrim", trimNode)
		c.graph.Connect(lastNode, trimNode, currentLabel, engine.StreamAudio)

		aptsNode := c.graph.AddFilter("asetpts", map[string]string{
			"expr": "PTS-STARTPTS",
		})
		aptsLabel := c.nextLabelFor("vapts", aptsNode)
		c.graph.Connect(trimNode, aptsNode, trimLabel, engine.StreamAudio)

		currentLabel = aptsLabel
		lastNode = aptsNode
	}

	// Apply volume.
	if cl.Volume() != 1.0 {
		volNode := c.graph.AddFilter("volume", map[string]string{
			"volume": fmt.Sprintf("%g", cl.Volume()),
		})
		volLabel := c.nextLabelFor("vavol", volNode)
		c.graph.Connect(lastNode, volNode, currentLabel, engine.StreamAudio)
		currentLabel = volLabel
		lastNode = volNode
	}

	// Apply composable audio effects from the clip's Effects() list.
	for _, e := range cl.Effects() {
		if e.Target() != effects.TargetAudio && e.Target() != effects.TargetBoth {
			continue
		}
		currentLabel, lastNode = c.compileAudioEffect(e, currentLabel, lastNode, "vafx")
	}

	return currentLabel, nil
}

// insertVideoGaps walks sorted video labels and inserts black gap clips
// wherever there is a time gap between the end of one clip and the start of
// the next. This ensures that the concat filter produces output where clips
// appear at their intended absolute positions instead of back-to-back.
//
// A gap before the first clip (i.e. first clip's StartAt > 0) is also filled,
// so Phase 2b's tpad is only needed for the single-clip case.
func (c *Compiler) insertVideoGaps(labels []clipLabel, cfg Config) []clipLabel {
	if len(labels) <= 1 {
		return labels
	}

	var result []clipLabel
	cursor := time.Duration(0)

	for _, cl := range labels {
		gap := cl.entry.StartAt - cursor
		if gap > 0 {
			gapLabel := c.compileGapClip(gap, cfg)
			result = append(result, gapLabel)
		}
		result = append(result, cl)
		cursor = cl.entry.StartAt + cl.entry.Clip.Duration()
	}

	return result
}

// compileGapClip creates a black video segment (with silence) of the given
// duration at the timeline resolution. It is used to fill time gaps between
// video clips so that concat-based positioning matches absolute StartAt times.
func (c *Compiler) compileGapClip(d time.Duration, cfg Config) clipLabel {
	colorNode := c.graph.AddFilter("color", map[string]string{
		"c": "black",
		"s": fmt.Sprintf("%dx%d", cfg.Width, cfg.Height),
		"d": formatSeconds(d),
		"r": fmt.Sprintf("%g", cfg.FPS),
	})
	colorLabel := c.nextLabelFor("gap", colorNode)

	// Normalize pixel format to yuv420p for concat compatibility.
	fmtNode := c.graph.AddFilter("format", map[string]string{
		"pix_fmts": "yuv420p",
	})
	fmtLabel := c.nextLabelFor("gapfmt", fmtNode)
	c.graph.Connect(colorNode, fmtNode, colorLabel, engine.StreamVideo)

	return clipLabel{
		video: fmtLabel,
		entry: Placement{
			StartAt: 0,
			Clip:    clip.NewColor("black", cfg.Width, cfg.Height).WithDuration(d),
		},
	}
}

// concatVideoClips concatenates multiple video clip outputs using the concat filter.
// If any clips have paired audio, the audio is concatenated alongside the video.
// Returns the final video label and (if applicable) the final audio label.
func (c *Compiler) concatVideoClips(labels []clipLabel) (string, string) {
	// Check if any clips have audio to concatenate.
	hasAudio := false
	for _, cl := range labels {
		if cl.audio != "" {
			hasAudio = true
			break
		}
	}

	audioFlag := "0"
	if hasAudio {
		audioFlag = "1"
	}

	concatNode := c.graph.AddFilter("concat", map[string]string{
		"n": fmt.Sprintf("%d", len(labels)),
		"v": "1",
		"a": audioFlag,
	})
	videoLabel := c.nextLabelFor("vconcat", concatNode)

	// For concat with audio, FFmpeg expects interleaved inputs:
	// [v0][a0][v1][a1]...concat=n=N:v=1:a=1[outv][outa]
	// We connect video first, then audio for each segment, in order.
	for _, cl := range labels {
		node := c.findFilterByLabel(cl.video)
		if node != nil {
			c.graph.Connect(node, concatNode, cl.video, engine.StreamVideo)
		}
		if hasAudio {
			if cl.audio != "" {
				anode := c.findFilterByLabel(cl.audio)
				if anode != nil {
					c.graph.Connect(anode, concatNode, cl.audio, engine.StreamAudio)
				}
			} else {
				// Clip has no audio — generate a silent placeholder so concat
				// input count stays balanced.
				silenceNode := c.graph.AddFilter("anullsrc", map[string]string{
					"r":  "48000",
					"cl": "stereo",
					"d":  formatSeconds(cl.entry.Clip.Duration()),
				})
				silenceLabel := c.nextLabelFor("asilence", silenceNode)
				c.graph.Connect(silenceNode, concatNode, silenceLabel, engine.StreamAudio)
			}
		}
	}

	audioLabel := ""
	if hasAudio {
		audioLabel = c.nextLabelFor("aconcat", concatNode)
	}

	return videoLabel, audioLabel
}

// mixAudio mixes multiple audio streams using the amix filter.
func (c *Compiler) mixAudio(labels []string) string {
	mixNode := c.graph.AddFilter("amix", map[string]string{
		"inputs":    fmt.Sprintf("%d", len(labels)),
		"duration":  "longest",
		"normalize": "0",
	})
	mixLabel := c.nextLabelFor("amix", mixNode)

	for _, label := range labels {
		node := c.findFilterByLabel(label)
		if node != nil {
			c.graph.Connect(node, mixNode, label, engine.StreamAudio)
		}
	}

	return mixLabel
}

// findFilterByLabel finds the node that produces the given label.
// It checks the labelMap first, then falls back to input references ("N:v"/"N:a")
// and output edge labels on filter nodes.
func (c *Compiler) findFilterByLabel(label string) *engine.Node {
	// Check the explicit label map first.
	if node, ok := c.labelMap[label]; ok {
		return node
	}

	// Check if it's a direct input reference like "0:v" or "1:a".
	var inputIdx int
	var streamChar string
	if n, _ := fmt.Sscanf(label, "%d:%s", &inputIdx, &streamChar); n == 2 {
		inputs := c.graph.Inputs()
		if inputIdx >= 0 && inputIdx < len(inputs) {
			return inputs[inputIdx]
		}
	}

	// Search filter nodes for one with this output label.
	for _, node := range c.graph.Nodes() {
		for _, edge := range node.Outputs {
			if edge.Label == label {
				return node
			}
		}
	}

	return nil
}

// compileAudioEffect emits the filter nodes for a single audio effect.
// Most effects produce a single filter node, but atempo needs special
// handling: FFmpeg's atempo filter only accepts factors in [0.5, 2.0],
// so higher or lower factors must be decomposed into a chain.
//
// Returns the updated (currentLabel, lastNode) after the effect chain.
func (c *Compiler) compileAudioEffect(e effects.Effect, currentLabel string, lastNode *engine.Node, labelPrefix string) (string, *engine.Node) {
	if e.FilterName() == "atempo" {
		return c.compileAtempoChain(e, currentLabel, lastNode, labelPrefix)
	}
	filterNode := c.graph.AddFilter(e.FilterName(), e.FilterParams())
	filterLabel := c.nextLabelFor(labelPrefix, filterNode)
	c.graph.Connect(lastNode, filterNode, currentLabel, engine.StreamAudio)
	return filterLabel, filterNode
}

// compileAtempoChain decomposes an atempo effect into a chain of atempo
// filters, each with a factor within FFmpeg's supported range of [0.5, 2.0].
//
// For example, a 4x speedup becomes: atempo=2.0 -> atempo=2.0
// A 6x speedup becomes: atempo=2.0 -> atempo=2.0 -> atempo=1.5
// A 0.25x slowdown becomes: atempo=0.5 -> atempo=0.5
func (c *Compiler) compileAtempoChain(e effects.Effect, currentLabel string, lastNode *engine.Node, labelPrefix string) (string, *engine.Node) {
	params := e.FilterParams()
	// Parse the tempo value from the params. The AudioSpeedEffect stores it
	// as "tempo" key. If we can't find it, fall back to single filter.
	tempoStr, ok := params["tempo"]
	if !ok {
		// Not a standard atempo effect — emit as-is.
		filterNode := c.graph.AddFilter(e.FilterName(), e.FilterParams())
		filterLabel := c.nextLabelFor(labelPrefix, filterNode)
		c.graph.Connect(lastNode, filterNode, currentLabel, engine.StreamAudio)
		return filterLabel, filterNode
	}

	var factor float64
	if _, err := fmt.Sscanf(tempoStr, "%f", &factor); err != nil || factor <= 0 {
		// Can't parse — emit as-is.
		filterNode := c.graph.AddFilter(e.FilterName(), e.FilterParams())
		filterLabel := c.nextLabelFor(labelPrefix, filterNode)
		c.graph.Connect(lastNode, filterNode, currentLabel, engine.StreamAudio)
		return filterLabel, filterNode
	}

	// Decompose into chain of atempo filters each within [0.5, 2.0].
	factors := decomposeAtempo(factor)
	for _, f := range factors {
		node := c.graph.AddFilter("atempo", map[string]string{
			"tempo": formatFloat(f),
		})
		label := c.nextLabelFor(labelPrefix, node)
		c.graph.Connect(lastNode, node, currentLabel, engine.StreamAudio)
		currentLabel = label
		lastNode = node
	}

	return currentLabel, lastNode
}

// decomposeAtempo breaks a tempo factor into a slice of factors each within
// FFmpeg's atempo range of [0.5, 2.0].
func decomposeAtempo(factor float64) []float64 {
	if factor >= 0.5 && factor <= 2.0 {
		return []float64{factor}
	}

	var factors []float64
	remaining := factor

	if remaining > 2.0 {
		for remaining > 2.0 {
			factors = append(factors, 2.0)
			remaining /= 2.0
		}
		if remaining > 1.0001 || remaining < 0.9999 { // avoid no-op 1.0
			factors = append(factors, remaining)
		}
	} else {
		// remaining < 0.5
		for remaining < 0.5 {
			factors = append(factors, 0.5)
			remaining /= 0.5
		}
		if remaining > 1.0001 || remaining < 0.9999 {
			factors = append(factors, remaining)
		}
	}

	return factors
}

// formatFloat formats a float64 for FFmpeg filter parameters, trimming
// trailing zeros for readability.
func formatFloat(f float64) string {
	s := fmt.Sprintf("%.6f", f)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}

// formatSeconds converts a Duration to a decimal seconds string for FFmpeg.
func formatSeconds(d time.Duration) string {
	return fmt.Sprintf("%.3f", d.Seconds())
}
