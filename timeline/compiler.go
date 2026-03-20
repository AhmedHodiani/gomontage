package timeline

import (
	"fmt"
	"strings"
	"time"

	"github.com/ahmedhodiani/gomontage/clip"
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

// Compile transforms the timeline into an FFmpeg filter graph.
// The resulting graph can be built into a Command via engine.BuildCommand.
func (c *Compiler) Compile(outputPath string, outputParams map[string]string) (*engine.Graph, error) {
	if err := c.timeline.Validate(); err != nil {
		return nil, fmt.Errorf("timeline validation failed: %w", err)
	}

	cfg := c.timeline.Config()

	// Phase 1: Process each video track — trim, apply per-clip effects.
	// Also compile the audio stream from video clips that have audio,
	// so that video and its accompanying audio can be concatenated together.
	var videoLabels []clipLabel
	for _, track := range c.timeline.videoTracks {
		for _, entry := range track.Entries() {
			labels, err := c.compileVideoEntry(entry, cfg)
			if err != nil {
				return nil, fmt.Errorf("track %q: %w", track.Name(), err)
			}
			// For video clips that have audio, compile the audio stream too.
			for i, cl := range labels {
				if cl.entry.Clip.HasAudio() && cl.entry.Clip.SourcePath() != "" {
					audioLabel, err := c.compileAudioFromVideoEntry(cl.entry)
					if err != nil {
						return nil, fmt.Errorf("track %q audio: %w", track.Name(), err)
					}
					labels[i].audio = audioLabel
				}
			}
			videoLabels = append(videoLabels, labels...)
		}
	}

	// Phase 2: Concatenate video clips (and their paired audio) if there are multiple.
	// Audio from video clips must be concatenated in sync with the video, not mixed.
	finalVideoLabel := ""
	finalVideoAudioLabel := ""
	if len(videoLabels) == 1 {
		finalVideoLabel = videoLabels[0].video
		finalVideoAudioLabel = videoLabels[0].audio
	} else if len(videoLabels) > 1 {
		finalVideoLabel, finalVideoAudioLabel = c.concatVideoClips(videoLabels)
	}

	// Phase 2b: Apply timeline delay to the final video/audio.
	// For a single clip, delay if its StartAt > 0.
	// For a concatenated sequence, delay only if the first clip starts after 0
	// (meaning the entire sequence is shifted forward on the timeline).
	// Per-clip StartAt within a sequence is handled by concat ordering, not tpad.
	if len(videoLabels) > 0 {
		firstStartAt := videoLabels[0].entry.StartAt
		if firstStartAt > 0 {
			// Apply tpad to delay video.
			tpadNode := c.graph.AddFilter("tpad", map[string]string{
				"start_duration": formatSeconds(firstStartAt),
				"color":          "black",
			})
			tpadLabel := c.nextLabelFor("vpad", tpadNode)
			c.graph.Connect(
				c.findFilterByLabel(finalVideoLabel),
				tpadNode,
				finalVideoLabel,
				engine.StreamVideo,
			)
			finalVideoLabel = tpadLabel

			// Apply adelay to paired audio from video.
			if finalVideoAudioLabel != "" {
				delayNode := c.graph.AddFilter("adelay", map[string]string{
					"delays": fmt.Sprintf("%d|%d", firstStartAt.Milliseconds(), firstStartAt.Milliseconds()),
				})
				delayLabel := c.nextLabelFor("vadel", delayNode)
				c.graph.Connect(
					c.findFilterByLabel(finalVideoAudioLabel),
					delayNode,
					finalVideoAudioLabel,
					engine.StreamAudio,
				)
				finalVideoAudioLabel = delayLabel
			}
		}
	}

	// Phase 3: Process independent audio tracks (background music, narration, etc.).
	var audioLabels []string
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

	// Add the concatenated audio from video tracks as one of the audio streams
	// to be mixed with independent audio tracks.
	if finalVideoAudioLabel != "" {
		audioLabels = append([]string{finalVideoAudioLabel}, audioLabels...)
	}

	// Phase 4: Mix audio if there are multiple independent audio streams.
	finalAudioLabel := ""
	if len(audioLabels) == 1 {
		finalAudioLabel = audioLabels[0]
	} else if len(audioLabels) > 1 {
		finalAudioLabel = c.mixAudio(audioLabels)
	}

	// Phase 5: Create output node.
	if outputParams == nil {
		outputParams = make(map[string]string)
	}
	output := c.graph.AddOutput(outputPath, outputParams)

	// Map final video and audio to output.
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

	return currentLabel, nil
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

// formatSeconds converts a Duration to a decimal seconds string for FFmpeg.
func formatSeconds(d time.Duration) string {
	return fmt.Sprintf("%.3f", d.Seconds())
}
