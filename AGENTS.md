# AGENTS.md — Gomontage

Programmatic video editing framework for Go. Module: `github.com/ahmedhodiani/gomontage`.
Users write Go `main.go` scripts importing this library to describe video edits; the framework wraps FFmpeg CLI.

## Build & Test Commands

```bash
# Run all tests
go test ./...

# Run tests for a single package
go test ./timeline/
go test ./clip/
go test ./engine/

# Run a single test by name (-run accepts a regex)
go test ./timeline/ -run TestDryRun_TrimFromZero
go test ./clip/ -run TestVideoClip_Trim

# Verbose output
go test -v ./timeline/ -run TestDryRun

# Clear test cache and re-run
go clean -testcache && go test ./...

# Build the CLI binary
go build ./cmd/gomontage/

# Vet (no linter config exists; use go vet)
go vet ./...
```

There is no Makefile, no CI config, no `.golangci.yml`. Use `go vet` and `gofmt`.

## Architecture — 4 Layers

1. **engine/** — FFmpeg interface: ffprobe wrapper, filter graph DAG, command builder, process runner
2. **clip/** — Immutable media types: VideoClip, AudioClip, ImageClip, TextClip, ColorClip
3. **timeline/** — Track-based timeline: VideoTrack/AudioTrack, Compiler (timeline → filter graph), Export/DryRun
4. **cmd/gomontage/** — Cobra CLI: init, run, probe, validate, docs

Supporting packages: `cuts/` (9 transition types), `effects/` (composable effects), `export/` (profiles/presets), `project/` (config YAML + scaffold), `docs/` (markdown generator).

Users interact with layer 3+ only. The root `gomontage.go` is a convenience facade over `timeline`.

## Code Style

### Imports

Three groups separated by blank lines: (1) stdlib, (2) third-party, (3) internal.

```go
import (
    "fmt"
    "strings"
    "time"

    "github.com/spf13/cobra"

    "github.com/ahmedhodiani/gomontage/clip"
    "github.com/ahmedhodiani/gomontage/engine"
)
```

Single-import files use the inline form: `import "time"`.

### Formatting

Standard `gofmt`. No custom formatter config.

### Naming

- **Acronyms stay uppercase**: `FPS`, `CRF`, `ID`, `PTS`, `URL` (e.g., `cfg.FPS`, `node.ID`)
- **Enum types**: `type Type int` with `const ( TypeVideo Type = iota; TypeAudio; ... )`
- **Enum prefix matches type**: `TypeVideo`, `TrackTypeVideo`, `StreamVideo`, `TargetVideo`, `NodeInput`, `TransitionHardCut`
- **Constructors**: `New<Type>()` or `New<Type>WithX()` (e.g., `NewVideo()`, `NewVideoWithDuration()`, `NewCompiler()`, `NewProfile()`)
- **Builder methods**: `With<Property>()` returning `*T` for chaining (e.g., `WithVolume()`, `WithFadeIn()`, `WithName()`)
- **Factory functions for value types**: short name matching the concept (e.g., `Hard()`, `LCut()`, `FadeIn()`, `At()`)
- **Unexported fields**: all struct fields are unexported; access via exported methods

### Types

- **No `map[string]interface{}`** — every FFmpeg option is `map[string]string` with typed Go structs
- **Interfaces are small**: `Clip` (13 methods), `Transition` (2 methods), `Effect` (4 methods)
- **Compile-time interface checks** at package level:
  ```go
  var _ timeline.Transition = (*HardCut)(nil)
  ```
- **Config structs** use typed fields, not option maps: `Config{Width: 1920, Height: 1080, FPS: 30}`

### Immutability

All clip transforms return new instances, never mutate. Pattern:

```go
func (c *VideoClip) Trim(start, end time.Duration) *VideoClip {
    n := &VideoClip{Base: *c.base()}  // shallow copy via base()
    n.trimStart = start
    n.trimEnd = end
    n.duration = end - start
    n.trimmed = true
    return n
}
```

### Error Handling

- Wrap with `fmt.Errorf("context: %w", err)` — always use `%w` for wrapping
- Add context at each level: `fmt.Errorf("track %q: %w", track.Name(), err)`
- No custom error types; all errors are wrapped stdlib errors
- Return `error` as the last return value
- CLI commands use `RunE` (error-returning) and print to stderr on failure

### Doc Comments

- Every exported symbol has a doc comment starting with the symbol name
- Every package has a `doc.go` with a package-level doc comment (except `docs/`)
- Doc comments include `// Example:` blocks with indented code where helpful
- Inline comments explain "why", not "what"

### Package-Level Organization

Each package follows: `doc.go` (package docs) → main source files → `*_test.go`.

## Testing Conventions

### Framework

Standard library `testing` only. No testify, no gomock, no third-party assertion libraries.

### Assertion Style

```go
if got != want {
    t.Errorf("FunctionName(%v) = %v, want %v", input, got, want)
}
```

Use `t.Fatalf` only for fatal precondition failures. Use `t.Errorf` for assertions that should continue.

### Test Patterns

- **Table-driven tests** for functions with multiple input/output pairs
- **Immutability tests** verify transforms don't mutate originals
- **Interface compliance tests** verify all types implement their interface
- **Regression tests** with comments: `// Regression test: <description of the bug>`
- **`strings.Contains`** on `cmd.String()` for FFmpeg command output validation
- **`t.TempDir()`** for filesystem tests (scaffold, docs generation)
- **White-box tests** (same package, not `_test` suffix) — tests access unexported functions

### Mocks

Hand-rolled mocks only. Example: `mockTransition` in `timeline/timeline_test.go`. No mocking frameworks.

### Test Fixtures

All test data is created inline. Use `NewVideoWithDuration()` / `NewAudioWithDuration()` to avoid needing real media files. No external fixture files.

## Version Management

Single source of truth in `version.go`:

```go
const Version = "0.1.0"
const GoVersion = "1.25"
```

All references (CLI `--version`, scaffolded `go.mod`, docs) derive from these constants. Update `Version` before tagging a release.

## Commit Style

Conventional commits: `feat:`, `fix:`, `refactor:`, `chore:`, `docs:`, `test:`.
Scope optional: `feat(timeline):`, `fix(clip):`. Keep commits small and focused.

## Key Design Principles

1. **Lazy evaluation** — nothing touches FFmpeg until `.Export()` is called; `DryRun()` returns the command without executing
2. **Immutability** — all clip transforms return new clips
3. **Type safety** — typed structs for all options, no untyped maps
4. **Track-based model** — named video/audio tracks with clips placed at specific times, like a real NLE
5. **FFmpeg CLI wrapper** — shells out to `ffmpeg`/`ffprobe` binaries on PATH (not cgo)
