package engine

import "fmt"

// NodeType identifies what kind of node exists in the filter graph.
type NodeType int

const (
	// NodeInput represents a media file input (-i flag).
	NodeInput NodeType = iota
	// NodeFilter represents an FFmpeg filter (trim, overlay, etc.).
	NodeFilter
	// NodeOutput represents the output file.
	NodeOutput
)

// StreamType identifies whether a stream is video or audio.
type StreamType int

const (
	// StreamVideo is a video stream.
	StreamVideo StreamType = iota
	// StreamAudio is an audio stream.
	StreamAudio
)

// String returns "v" for video and "a" for audio.
func (s StreamType) String() string {
	if s == StreamAudio {
		return "a"
	}
	return "v"
}

// Node is a single node in the FFmpeg filter graph DAG.
// It can represent an input file, a filter operation, or an output.
type Node struct {
	// ID is the unique identifier for this node within the graph.
	ID string

	// Type identifies what kind of node this is.
	Type NodeType

	// Name is the filter name (e.g. "trim", "overlay", "amix") for filter nodes,
	// or the file path for input/output nodes.
	Name string

	// Params holds the typed parameters for this node.
	Params map[string]string

	// Inputs are the edges coming into this node (other nodes feeding into this one).
	Inputs []*Edge

	// Outputs are the edges going out of this node.
	Outputs []*Edge
}

// Edge connects two nodes in the graph. It represents a stream flowing
// from one node to another, carrying a label (e.g. "[0:v]", "[trimmed]").
type Edge struct {
	// From is the source node.
	From *Node

	// To is the destination node.
	To *Node

	// Label is the FFmpeg stream label (e.g. "0:v", "trimmed", "out_v").
	Label string

	// Stream identifies whether this edge carries video or audio.
	Stream StreamType
}

// Graph is a directed acyclic graph representing an FFmpeg filter chain.
// Nodes are inputs, filters, and outputs. Edges are streams flowing between them.
type Graph struct {
	// nodes stores all nodes by ID.
	nodes map[string]*Node

	// inputs is the ordered list of input nodes (order matters for -i indexing).
	inputs []*Node

	// outputs is the list of output nodes.
	outputs []*Node

	// nodeCounter is used to generate unique node IDs.
	nodeCounter int
}

// NewGraph creates an empty filter graph.
func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[string]*Node),
	}
}

// nextID generates a unique node ID.
func (g *Graph) nextID(prefix string) string {
	g.nodeCounter++
	return fmt.Sprintf("%s_%d", prefix, g.nodeCounter)
}

// AddInput adds an input file to the graph and returns the node.
// The order of AddInput calls determines the input index (0, 1, 2...).
func (g *Graph) AddInput(path string) *Node {
	id := g.nextID("in")
	node := &Node{
		ID:     id,
		Type:   NodeInput,
		Name:   path,
		Params: make(map[string]string),
	}
	g.nodes[id] = node
	g.inputs = append(g.inputs, node)
	return node
}

// AddFilter adds a filter node to the graph and returns it.
// The filter is not connected to anything yet — use Connect to wire it up.
func (g *Graph) AddFilter(name string, params map[string]string) *Node {
	id := g.nextID("f")
	if params == nil {
		params = make(map[string]string)
	}
	node := &Node{
		ID:     id,
		Type:   NodeFilter,
		Name:   name,
		Params: params,
	}
	g.nodes[id] = node
	return node
}

// AddOutput adds an output file node to the graph and returns it.
func (g *Graph) AddOutput(path string, params map[string]string) *Node {
	id := g.nextID("out")
	if params == nil {
		params = make(map[string]string)
	}
	node := &Node{
		ID:     id,
		Type:   NodeOutput,
		Name:   path,
		Params: params,
	}
	g.nodes[id] = node
	g.outputs = append(g.outputs, node)
	return node
}

// Connect creates an edge from src to dst with the given stream label and type.
func (g *Graph) Connect(src, dst *Node, label string, stream StreamType) *Edge {
	edge := &Edge{
		From:   src,
		To:     dst,
		Label:  label,
		Stream: stream,
	}
	src.Outputs = append(src.Outputs, edge)
	dst.Inputs = append(dst.Inputs, edge)
	return edge
}

// InputIndex returns the 0-based index of an input node, which corresponds
// to its position in the FFmpeg command's -i arguments.
// Returns -1 if the node is not an input.
func (g *Graph) InputIndex(node *Node) int {
	for i, n := range g.inputs {
		if n.ID == node.ID {
			return i
		}
	}
	return -1
}

// Inputs returns all input nodes in order.
func (g *Graph) Inputs() []*Node {
	return g.inputs
}

// Outputs returns all output nodes.
func (g *Graph) Outputs() []*Node {
	return g.outputs
}

// Nodes returns all nodes in the graph.
func (g *Graph) Nodes() map[string]*Node {
	return g.nodes
}
