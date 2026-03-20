// Package engine provides the low-level FFmpeg interface for Gomontage.
//
// This package is internal to the framework. Users of Gomontage interact
// with higher-level abstractions (clip, timeline, cuts, effects) that
// compile down to engine operations.
//
// The engine handles:
//   - Media probing via ffprobe (metadata extraction)
//   - Filter graph construction as a typed DAG
//   - FFmpeg command building from the filter graph
//   - Process execution with progress reporting
package engine
