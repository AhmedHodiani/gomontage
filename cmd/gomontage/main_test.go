package main

import (
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{500 * time.Millisecond, "0.5s"},
		{2 * time.Second, "2.0s"},
		{90 * time.Second, "1m 30.0s"},
		{3661 * time.Second, "1h 1m 1.0s"},
	}

	for _, tt := range tests {
		got := formatDuration(tt.d)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{500, "500 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1048576, "1.00 MB"},
		{1073741824, "1.00 GB"},
	}

	for _, tt := range tests {
		got := formatBytes(tt.bytes)
		if got != tt.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}

func TestFormatBitrate(t *testing.T) {
	tests := []struct {
		bps  int64
		want string
	}{
		{128000, "128 kbps"},
		{1500000, "1.5 Mbps"},
		{5000000, "5.0 Mbps"},
	}

	for _, tt := range tests {
		got := formatBitrate(tt.bps)
		if got != tt.want {
			t.Errorf("formatBitrate(%d) = %q, want %q", tt.bps, got, tt.want)
		}
	}
}

func TestCommandsRegistered(t *testing.T) {
	// Verify that all command constructors don't panic and return valid commands.
	commands := []struct {
		name string
		fn   func() *cobra.Command
	}{
		{"init", initCmd},
		{"run", runCmd},
		{"probe", probeCmd},
		{"validate", validateCmd},
		{"docs", docsCmd},
	}

	for _, tc := range commands {
		cmd := tc.fn()
		if cmd == nil {
			t.Errorf("%s command returned nil", tc.name)
			continue
		}
		if cmd.Use == "" {
			t.Errorf("%s command has empty Use field", tc.name)
		}
		if cmd.Short == "" {
			t.Errorf("%s command has empty Short field", tc.name)
		}
		if cmd.RunE == nil {
			t.Errorf("%s command has nil RunE", tc.name)
		}
	}
}

func TestDocsCommandDefaultOutput(t *testing.T) {
	cmd := docsCmd()
	output, err := cmd.Flags().GetString("output")
	if err != nil {
		t.Fatalf("could not get output flag: %v", err)
	}
	if output != "docs" {
		t.Errorf("expected default output to be %q, got %q", "docs", output)
	}
}

func TestInitCommandRequiresArg(t *testing.T) {
	cmd := initCmd()
	// cobra.ExactArgs(1) should reject 0 args
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("expected init command to reject zero arguments")
	}
}

func TestProbeCommandRequiresArg(t *testing.T) {
	cmd := probeCmd()
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("expected probe command to reject zero arguments")
	}
}
