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
	var videoLabels []clipLabel
	for _, track := range c.timeline.videoTracks {
		for _, entry := range track.Entries() {
			labels, err := c.compileVideoEntry(entry, cfg)
			if err != nil {
				return nil, fmt.Errorf("track %q: %w", track.Name(), err)
			}
			videoLabels = append(videoLabels, labels...)
		}
	}

	// Phase 2: Concatenate video clips if there are multiple.
	finalVideoLabel := ""
	if len(videoLabels) == 1 {
		finalVideoLabel = videoLabels[0].video
	} else if len(videoLabels) > 1 {
		finalVideoLabel = c.concatVideoClips(videoLabels)
	}

	// Phase 3: Process audio tracks.
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

	// Also collect audio from video tracks (video clips that have audio).
	for _, track := range c.timeline.videoTracks {
		for _, entry := range track.Entries() {
			if entry.Clip.HasAudio() && entry.Clip.SourcePath() != "" {
				label, err := c.compileAudioFromVideoEntry(entry)
				if err != nil {
					return nil, fmt.Errorf("video track %q audio: %w", track.Name(), err)
				}
				if label != "" {
					audioLabels = append(audioLabels, label)
				}
			}
		}
	}

	// Phase 4: Mix audio if there are multiple audio streams.
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

	if path == "" {
		// Generated clips (color, text) — handled differently.
		return c.compileGeneratedClip(entry, cfg)
	}

	inputNode := c.getOrAddInput(path)
	inputIdx := c.graph.InputIndex(inputNode)

	// Start with the raw video stream.
	currentLabel := fmt.Sprintf("%d:v", inputIdx)
	var lastNode *engine.Node = inputNode

	// Apply trim if needed.
	if cl.TrimStart() > 0 || cl.TrimEnd() < cl.Duration()+cl.TrimStart() {
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

	// Apply scale if the clip has specific dimensions.
	if cl.Width() > 0 && cl.Height() > 0 {
		scaleNode := c.graph.AddFilter("scale", map[string]string{
			"w": fmt.Sprintf("%d", cl.Width()),
			"h": fmt.Sprintf("%d", cl.Height()),
		})
		scaleLabel := c.nextLabelFor("vscale", scaleNode)
		c.graph.Connect(lastNode, scaleNode, currentLabel, engine.StreamVideo)
		currentLabel = scaleLabel
		lastNode = scaleNode
	}

	// Apply fade in.
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

	// Apply fade out.
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
func (c *Compiler) compileGeneratedClip(entry Placement, cfg Config) ([]clipLabel, error) {
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
		return []clipLabel{{video: label, entry: entry}}, nil

	case *clip.TextClip:
		// Text clips use the drawtext filter on a color background.
		bgNode := c.graph.AddFilter("color", map[string]string{
			"c": "black@0.0", // Transparent background.
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
		label := c.nextLabelFor("gen", textNode)
		c.graph.Connect(bgNode, textNode, bgLabel, engine.StreamVideo)

		return []clipLabel{{video: label, entry: entry}}, nil

	default:
		return nil, fmt.Errorf("unsupported generated clip type: %T", cl)
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
	if cl.TrimStart() > 0 || cl.TrimEnd() < cl.Duration()+cl.TrimStart() {
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
	if cl.TrimStart() > 0 || cl.TrimEnd() < cl.Duration()+cl.TrimStart() {
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
func (c *Compiler) concatVideoClips(labels []clipLabel) string {
	concatNode := c.graph.AddFilter("concat", map[string]string{
		"n": fmt.Sprintf("%d", len(labels)),
		"v": "1",
		"a": "0",
	})
	concatLabel := c.nextLabelFor("vconcat", concatNode)

	for _, cl := range labels {
		// Find the node that produced this label.
		node := c.findFilterByLabel(cl.video)
		if node != nil {
			c.graph.Connect(node, concatNode, cl.video, engine.StreamVideo)
		}
	}

	return concatLabel
}

// mixAudio mixes multiple audio streams using the amix filter.
func (c *Compiler) mixAudio(labels []string) string {
	mixNode := c.graph.AddFilter("amix", map[string]string{
		"inputs":   fmt.Sprintf("%d", len(labels)),
		"duration": "longest",
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
