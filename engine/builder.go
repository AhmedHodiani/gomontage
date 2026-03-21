package engine

import (
	"fmt"
	"sort"
	"strings"
)

// Command represents a fully-built FFmpeg command ready for execution.
type Command struct {
	// Binary is the FFmpeg executable name or path.
	Binary string

	// Args is the complete argument list.
	Args []string
}

// String returns the command as a shell-friendly string for debugging.
func (c *Command) String() string {
	parts := make([]string, 0, len(c.Args)+1)
	parts = append(parts, c.Binary)
	for _, arg := range c.Args {
		if strings.Contains(arg, " ") || strings.Contains(arg, ";") || strings.Contains(arg, "[") {
			parts = append(parts, fmt.Sprintf("%q", arg))
		} else {
			parts = append(parts, arg)
		}
	}
	return strings.Join(parts, " ")
}

// BuildCommand compiles a filter graph into an FFmpeg Command.
//
// The generated command follows this structure:
//
//	ffmpeg -y [global opts] -i input1 -i input2 ... -filter_complex "..." [output opts] output
func BuildCommand(g *Graph) (*Command, error) {
	if len(g.inputs) == 0 {
		return nil, fmt.Errorf("graph has no inputs")
	}
	if len(g.outputs) == 0 {
		return nil, fmt.Errorf("graph has no outputs")
	}

	cmd := &Command{Binary: "ffmpeg"}
	args := []string{"-y"} // Overwrite output without asking.

	// Add inputs in order.
	for _, input := range g.inputs {
		// Add input-level options before the -i flag.
		inputArgs := sortedParams(input.Params)
		args = append(args, inputArgs...)
		args = append(args, "-i", input.Name)
	}

	// Build filter_complex string from filter nodes.
	filterStr := buildFilterComplex(g)
	if filterStr != "" {
		args = append(args, "-filter_complex", filterStr)
	}

	// Add outputs.
	for _, output := range g.outputs {
		// Determine which streams map to this output.
		for _, edge := range output.Inputs {
			if edge.From.Type == NodeFilter {
				// Map from a filter graph output label.
				args = append(args, "-map", fmt.Sprintf("[%s]", edge.Label))
			} else if edge.From.Type == NodeInput {
				// Map directly from input stream.
				idx := g.InputIndex(edge.From)
				if idx >= 0 {
					args = append(args, "-map", fmt.Sprintf("%d:%s", idx, edge.Stream))
				}
			}
		}

		// Add output-level params.
		outputArgs := sortedParams(output.Params)
		args = append(args, outputArgs...)
		args = append(args, output.Name)
	}

	cmd.Args = args
	return cmd, nil
}

// buildFilterComplex generates the -filter_complex string from all filter nodes.
func buildFilterComplex(g *Graph) string {
	// Collect all filter nodes in dependency order.
	filters := topologicalSort(g)
	if len(filters) == 0 {
		return ""
	}

	var parts []string
	for _, node := range filters {
		var sb strings.Builder

		// Write input labels.
		for _, edge := range node.Inputs {
			if edge.From.Type == NodeInput {
				idx := g.InputIndex(edge.From)
				sb.WriteString(fmt.Sprintf("[%d:%s]", idx, edge.Stream))
			} else {
				sb.WriteString(fmt.Sprintf("[%s]", edge.Label))
			}
		}

		// Write filter name and params.
		sb.WriteString(node.Name)
		params := sortedParams(node.Params)
		if len(params) > 0 {
			sb.WriteString("=")
			// Join key=value pairs.
			kvPairs := make([]string, 0, len(params)/2)
			for i := 0; i < len(params); i += 2 {
				kvPairs = append(kvPairs, fmt.Sprintf("%s=%s", params[i], params[i+1]))
			}
			sb.WriteString(strings.Join(kvPairs, ":"))
		}

		// Write output labels. Sort video before audio so that multi-output
		// filters like concat (which produce video pads first, then audio pads)
		// get their labels mapped to the correct output pads.
		sortedOutputs := make([]*Edge, len(node.Outputs))
		copy(sortedOutputs, node.Outputs)
		sort.SliceStable(sortedOutputs, func(i, j int) bool {
			return sortedOutputs[i].Stream < sortedOutputs[j].Stream // StreamVideo(0) before StreamAudio(1)
		})
		for _, edge := range sortedOutputs {
			sb.WriteString(fmt.Sprintf("[%s]", edge.Label))
		}

		parts = append(parts, sb.String())
	}

	return strings.Join(parts, ";")
}

// topologicalSort returns filter nodes in dependency order (inputs before outputs).
func topologicalSort(g *Graph) []*Node {
	var filters []*Node
	for _, node := range g.nodes {
		if node.Type == NodeFilter {
			filters = append(filters, node)
		}
	}

	if len(filters) == 0 {
		return nil
	}

	// Kahn's algorithm.
	// Count incoming edges from other filter nodes.
	inDegree := make(map[string]int)
	for _, f := range filters {
		inDegree[f.ID] = 0
	}
	for _, f := range filters {
		for _, edge := range f.Inputs {
			if edge.From.Type == NodeFilter {
				inDegree[f.ID]++
			}
		}
	}

	// Start with nodes that have no filter dependencies.
	var queue []*Node
	for _, f := range filters {
		if inDegree[f.ID] == 0 {
			queue = append(queue, f)
		}
	}

	// Sort queue by ID for deterministic output.
	sort.Slice(queue, func(i, j int) bool { return queue[i].ID < queue[j].ID })

	var sorted []*Node
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		sorted = append(sorted, node)

		for _, edge := range node.Outputs {
			if edge.To.Type == NodeFilter {
				inDegree[edge.To.ID]--
				if inDegree[edge.To.ID] == 0 {
					queue = append(queue, edge.To)
				}
			}
		}
		// Re-sort for determinism after adding new nodes.
		sort.Slice(queue, func(i, j int) bool { return queue[i].ID < queue[j].ID })
	}

	return sorted
}

// sortedParams returns map entries as a flat slice of alternating key, value
// strings, sorted by key for deterministic output.
func sortedParams(params map[string]string) []string {
	if len(params) == 0 {
		return nil
	}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := make([]string, 0, len(keys)*2)
	for _, k := range keys {
		result = append(result, k, params[k])
	}
	return result
}
