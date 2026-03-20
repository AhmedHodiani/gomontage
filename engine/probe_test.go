package engine

import (
	"testing"
	"time"
)

func TestParseSeconds(t *testing.T) {
	tests := []struct {
		input   string
		want    time.Duration
		wantErr bool
	}{
		{"10.000000", 10 * time.Second, false},
		{"0.500000", 500 * time.Millisecond, false},
		{"123.456", 123*time.Second + 456*time.Millisecond, false},
		{"0", 0, false},
		{"", 0, true},
		{"not-a-number", 0, true},
	}

	for _, tt := range tests {
		got, err := parseSeconds(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseSeconds(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("parseSeconds(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParseFraction(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"30/1", 30.0},
		{"30000/1001", 29.97002997002997},
		{"24/1", 24.0},
		{"25.0", 25.0},
		{"", 0},
		{"invalid", 0},
	}

	for _, tt := range tests {
		got := parseFraction(tt.input)
		diff := got - tt.want
		if diff < 0 {
			diff = -diff
		}
		if diff > 0.001 {
			t.Errorf("parseFraction(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
