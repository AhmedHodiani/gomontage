package engine

import (
	"strings"
	"testing"
)

func TestBuildCommand_SimpleTranscode(t *testing.T) {
	g := NewGraph()

	input := g.AddInput("input.mp4")
	output := g.AddOutput("output.mp4", map[string]string{
		"-c:v": "libx264",
		"-c:a": "aac",
	})

	g.Connect(input, output, "0", StreamVideo)
	g.Connect(input, output, "0", StreamAudio)

	cmd, err := BuildCommand(g)
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}

	str := cmd.String()
	if !strings.Contains(str, "-i input.mp4") {
		t.Errorf("expected -i input.mp4, got: %s", str)
	}
	if !strings.Contains(str, "-c:v libx264") {
		t.Errorf("expected -c:v libx264, got: %s", str)
	}
	if !strings.Contains(str, "output.mp4") {
		t.Errorf("expected output.mp4, got: %s", str)
	}
}

func TestBuildCommand_WithFilter(t *testing.T) {
	g := NewGraph()

	input := g.AddInput("input.mp4")
	trim := g.AddFilter("trim", map[string]string{
		"start": "0",
		"end":   "10",
	})
	output := g.AddOutput("output.mp4", nil)

	g.Connect(input, trim, "0:v", StreamVideo)
	g.Connect(trim, output, "trimmed", StreamVideo)

	cmd, err := BuildCommand(g)
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}

	str := cmd.String()
	if !strings.Contains(str, "-filter_complex") {
		t.Errorf("expected -filter_complex, got: %s", str)
	}
	if !strings.Contains(str, "[0:v]trim=") {
		t.Errorf("expected [0:v]trim=, got: %s", str)
	}
	if !strings.Contains(str, "[trimmed]") {
		t.Errorf("expected [trimmed] output label, got: %s", str)
	}
	if !strings.Contains(str, "-map") {
		t.Errorf("expected -map, got: %s", str)
	}
}

func TestBuildCommand_MultipleInputs(t *testing.T) {
	g := NewGraph()

	in1 := g.AddInput("video1.mp4")
	in2 := g.AddInput("video2.mp4")

	overlay := g.AddFilter("overlay", map[string]string{
		"x": "0",
		"y": "0",
	})

	output := g.AddOutput("output.mp4", nil)

	g.Connect(in1, overlay, "0:v", StreamVideo)
	g.Connect(in2, overlay, "1:v", StreamVideo)
	g.Connect(overlay, output, "overlaid", StreamVideo)

	cmd, err := BuildCommand(g)
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}

	str := cmd.String()
	if !strings.Contains(str, "-i video1.mp4") {
		t.Errorf("expected -i video1.mp4, got: %s", str)
	}
	if !strings.Contains(str, "-i video2.mp4") {
		t.Errorf("expected -i video2.mp4, got: %s", str)
	}
	if !strings.Contains(str, "[0:v][1:v]overlay") {
		t.Errorf("expected [0:v][1:v]overlay, got: %s", str)
	}
}

func TestBuildCommand_ChainedFilters(t *testing.T) {
	g := NewGraph()

	input := g.AddInput("input.mp4")
	trim := g.AddFilter("trim", map[string]string{
		"start": "5",
		"end":   "15",
	})
	setpts := g.AddFilter("setpts", map[string]string{
		"expr": "PTS-STARTPTS",
	})
	output := g.AddOutput("output.mp4", nil)

	g.Connect(input, trim, "0:v", StreamVideo)
	g.Connect(trim, setpts, "trimmed", StreamVideo)
	g.Connect(setpts, output, "final", StreamVideo)

	cmd, err := BuildCommand(g)
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}

	str := cmd.String()
	// Verify the filter chain has both filters in order.
	trimIdx := strings.Index(str, "trim=")
	setptsIdx := strings.Index(str, "setpts=")
	if trimIdx < 0 || setptsIdx < 0 {
		t.Fatalf("expected both trim and setpts in: %s", str)
	}
	if trimIdx > setptsIdx {
		t.Errorf("expected trim before setpts, got: %s", str)
	}
}

func TestBuildCommand_NoInputs(t *testing.T) {
	g := NewGraph()
	g.AddOutput("output.mp4", nil)

	_, err := BuildCommand(g)
	if err == nil {
		t.Error("expected error for graph with no inputs")
	}
}

func TestBuildCommand_NoOutputs(t *testing.T) {
	g := NewGraph()
	g.AddInput("input.mp4")

	_, err := BuildCommand(g)
	if err == nil {
		t.Error("expected error for graph with no outputs")
	}
}

func TestGraph_InputIndex(t *testing.T) {
	g := NewGraph()
	in1 := g.AddInput("a.mp4")
	in2 := g.AddInput("b.mp4")
	in3 := g.AddInput("c.mp4")
	filter := g.AddFilter("null", nil)

	if g.InputIndex(in1) != 0 {
		t.Errorf("expected index 0 for in1, got %d", g.InputIndex(in1))
	}
	if g.InputIndex(in2) != 1 {
		t.Errorf("expected index 1 for in2, got %d", g.InputIndex(in2))
	}
	if g.InputIndex(in3) != 2 {
		t.Errorf("expected index 2 for in3, got %d", g.InputIndex(in3))
	}
	if g.InputIndex(filter) != -1 {
		t.Errorf("expected index -1 for filter, got %d", g.InputIndex(filter))
	}
}
