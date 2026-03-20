package docs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "docs")

	if err := Generate(outDir); err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	// Guide templates should be copied.
	guideFiles := []string{
		"getting-started.md",
		"clips.md",
		"timeline.md",
		"effects.md",
		"export.md",
	}
	for _, name := range guideFiles {
		path := filepath.Join(outDir, name)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("expected guide file %s to exist: %v", name, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("expected guide file %s to be non-empty", name)
		}
	}

	// API reference pages should be generated.
	apiFiles := []string{
		"api-reference.md",
		"api-gomontage.md",
		"api-clip.md",
		"api-timeline.md",
		"api-effects.md",
		"api-export.md",
	}
	for _, name := range apiFiles {
		path := filepath.Join(outDir, name)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("expected API file %s to exist: %v", name, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("expected API file %s to be non-empty", name)
		}
	}
}

func TestGenerateCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "a", "b", "c")

	if err := Generate(nested); err != nil {
		t.Fatalf("Generate() failed to create nested directory: %v", err)
	}

	info, err := os.Stat(nested)
	if err != nil {
		t.Fatalf("nested directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %s to be a directory", nested)
	}
}

func TestGenerateOverwritesExistingFiles(t *testing.T) {
	dir := t.TempDir()

	// Write a dummy file that should be overwritten.
	clipsMd := filepath.Join(dir, "clips.md")
	if err := os.WriteFile(clipsMd, []byte("old content"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Generate(dir); err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	data, err := os.ReadFile(clipsMd)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "old content" {
		t.Error("expected clips.md to be overwritten, but it still has old content")
	}
}

// --- Guide template content tests ---

func TestGettingStartedDocContent(t *testing.T) {
	content := readTemplate(t, "getting-started.md")
	mustContain := []string{
		"# Getting Started",
		"gomontage init",
		"gomontage run",
		"main.go",
		"timeline.New",
	}
	for _, s := range mustContain {
		if !strings.Contains(content, s) {
			t.Errorf("getting-started doc missing expected content: %q", s)
		}
	}
}

func TestClipsDocContent(t *testing.T) {
	content := readTemplate(t, "clips.md")
	clipTypes := []string{"VideoClip", "AudioClip", "ImageClip", "TextClip", "ColorClip"}
	for _, ct := range clipTypes {
		if !strings.Contains(content, ct) {
			t.Errorf("clips doc missing clip type: %s", ct)
		}
	}
	if !strings.Contains(content, "Immutability") {
		t.Error("clips doc missing immutability section")
	}
}

func TestEffectsDocContent(t *testing.T) {
	content := readTemplate(t, "effects.md")
	effects := []string{"FadeIn", "FadeOut", "SpeedUp", "Volume", "Normalize", "AudioSpeed"}
	for _, e := range effects {
		if !strings.Contains(content, e) {
			t.Errorf("effects doc missing effect: %s", e)
		}
	}
}

func TestExportDocContent(t *testing.T) {
	content := readTemplate(t, "export.md")
	presets := []string{"YouTube1080p", "YouTube4K", "Reel", "ProRes", "H265", "GIF", "MP3", "WAV", "FLAC"}
	for _, p := range presets {
		if !strings.Contains(content, p) {
			t.Errorf("export doc missing preset: %s", p)
		}
	}
	if !strings.Contains(content, "NewProfile") {
		t.Error("export doc missing custom profile builder section")
	}
}

// --- API reference content tests ---

func TestAPIReferenceIndex(t *testing.T) {
	dir := t.TempDir()
	if err := Generate(dir); err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	content := readFile(t, filepath.Join(dir, "api-reference.md"))

	// Should list all documented packages.
	pkgs := []string{"gomontage", "clip", "timeline", "effects", "export"}
	for _, pkg := range pkgs {
		if !strings.Contains(content, pkg) {
			t.Errorf("api-reference index missing package: %s", pkg)
		}
	}

	// Should link to per-package files.
	for _, pkg := range pkgs {
		filename := apiFilename(pkg)
		if !strings.Contains(content, filename) {
			t.Errorf("api-reference index missing link to %s", filename)
		}
	}

	// Should NOT include engine (internal).
	if strings.Contains(content, "api-engine.md") {
		t.Error("api-reference index should not include engine package")
	}
}

func TestAPIClipPage(t *testing.T) {
	dir := t.TempDir()
	if err := Generate(dir); err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	content := readFile(t, filepath.Join(dir, "api-clip.md"))

	// Should contain all clip types.
	types := []string{"VideoClip", "AudioClip", "ImageClip", "TextClip", "ColorClip"}
	for _, typ := range types {
		if !strings.Contains(content, typ) {
			t.Errorf("api-clip.md missing type: %s", typ)
		}
	}

	// Should contain the Clip interface.
	if !strings.Contains(content, "Clip") {
		t.Error("api-clip.md missing Clip interface")
	}

	// Should contain method signatures (auto-generated from source).
	methods := []string{"Trim", "WithVolume", "WithFadeIn", "WithFadeOut", "NewVideo", "NewAudio"}
	for _, m := range methods {
		if !strings.Contains(content, m) {
			t.Errorf("api-clip.md missing method: %s", m)
		}
	}

	// Should contain actual Go signatures with types.
	signatures := []string{"time.Duration", "*VideoClip", "*AudioClip"}
	for _, sig := range signatures {
		if !strings.Contains(content, sig) {
			t.Errorf("api-clip.md missing signature fragment: %s", sig)
		}
	}
}

func TestAPITimelinePage(t *testing.T) {
	dir := t.TempDir()
	if err := Generate(dir); err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	content := readFile(t, filepath.Join(dir, "api-timeline.md"))

	// Should contain timeline types.
	types := []string{"Timeline", "VideoTrack", "AudioTrack", "Config", "Placement"}
	for _, typ := range types {
		if !strings.Contains(content, typ) {
			t.Errorf("api-timeline.md missing type: %s", typ)
		}
	}

	// Should contain key methods.
	methods := []string{"AddVideoTrack", "AddAudioTrack", "Add", "AddSequence", "Export", "DryRun", "Validate"}
	for _, m := range methods {
		if !strings.Contains(content, m) {
			t.Errorf("api-timeline.md missing method: %s", m)
		}
	}
}

func TestAPIEffectsPage(t *testing.T) {
	dir := t.TempDir()
	if err := Generate(dir); err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	content := readFile(t, filepath.Join(dir, "api-effects.md"))

	// Should contain all effect types and factories.
	effects := []string{
		"FadeInEffect", "FadeOutEffect", "SpeedEffect",
		"AudioFadeInEffect", "AudioFadeOutEffect", "VolumeEffect",
		"NormalizeEffect", "AudioSpeedEffect",
		"FadeIn", "FadeOut", "SpeedUp", "SlowDown",
		"AudioFadeIn", "AudioFadeOut", "Volume", "Normalize", "AudioSpeed",
	}
	for _, e := range effects {
		if !strings.Contains(content, e) {
			t.Errorf("api-effects.md missing: %s", e)
		}
	}

	// Should contain the Effect interface.
	if !strings.Contains(content, "Effect") {
		t.Error("api-effects.md missing Effect interface")
	}
}

func TestAPIExportPage(t *testing.T) {
	dir := t.TempDir()
	if err := Generate(dir); err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	content := readFile(t, filepath.Join(dir, "api-export.md"))

	// Should contain profile type and builder.
	types := []string{"Profile", "ProfileBuilder"}
	for _, typ := range types {
		if !strings.Contains(content, typ) {
			t.Errorf("api-export.md missing type: %s", typ)
		}
	}

	// Should contain all presets.
	presets := []string{"YouTube1080p", "YouTube4K", "YouTube720p", "Reel", "ProRes", "H265", "GIF", "MP3", "WAV", "FLAC"}
	for _, p := range presets {
		if !strings.Contains(content, p) {
			t.Errorf("api-export.md missing preset: %s", p)
		}
	}

	// Should contain builder methods.
	builders := []string{"WithCodec", "WithCRF", "WithPreset", "WithAudioCodec", "Build"}
	for _, b := range builders {
		if !strings.Contains(content, b) {
			t.Errorf("api-export.md missing builder method: %s", b)
		}
	}

	// Profile struct fields should be documented.
	fields := []string{"VideoCodec", "AudioCodec", "CRF", "PixelFormat"}
	for _, f := range fields {
		if !strings.Contains(content, f) {
			t.Errorf("api-export.md missing field: %s", f)
		}
	}
}

func TestAPIGomontageRootPage(t *testing.T) {
	dir := t.TempDir()
	if err := Generate(dir); err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	content := readFile(t, filepath.Join(dir, "api-gomontage.md"))

	// Should contain root facade functions.
	funcs := []string{"NewTimeline", "At", "HD", "UHD", "Vertical", "Square", "Seconds", "Minutes"}
	for _, fn := range funcs {
		if !strings.Contains(content, fn) {
			t.Errorf("api-gomontage.md missing function: %s", fn)
		}
	}
}

// --- AST parser tests ---

func TestParsePackage(t *testing.T) {
	moduleRoot, err := findModuleRoot()
	if err != nil {
		t.Fatalf("findModuleRoot() error: %v", err)
	}

	pdoc, err := parsePackage(
		filepath.Join(moduleRoot, "clip"),
		"github.com/ahmedhodiani/gomontage/clip",
	)
	if err != nil {
		t.Fatalf("parsePackage() error: %v", err)
	}

	if pdoc.Name != "clip" {
		t.Errorf("expected package name 'clip', got %q", pdoc.Name)
	}

	// Should find the VideoClip type.
	var found bool
	for _, td := range pdoc.Types {
		if td.Name == "VideoClip" {
			found = true
			// Should have constructors.
			if len(td.Constructors) == 0 {
				t.Error("VideoClip should have constructors")
			}
			// Should have methods.
			if len(td.Methods) == 0 {
				t.Error("VideoClip should have methods")
			}
			// Check a known method exists with a signature.
			var hasTrim bool
			for _, m := range td.Methods {
				if m.Name == "Trim" {
					hasTrim = true
					if !strings.Contains(m.Signature, "time.Duration") {
						t.Errorf("Trim signature should contain time.Duration, got %q", m.Signature)
					}
					if m.Doc == "" {
						t.Error("Trim should have a doc comment")
					}
				}
			}
			if !hasTrim {
				t.Error("VideoClip should have a Trim method")
			}
			break
		}
	}
	if !found {
		t.Error("parsePackage should find VideoClip type")
	}
}

func TestParsePackageInterface(t *testing.T) {
	moduleRoot, err := findModuleRoot()
	if err != nil {
		t.Fatalf("findModuleRoot() error: %v", err)
	}

	pdoc, err := parsePackage(
		filepath.Join(moduleRoot, "clip"),
		"github.com/ahmedhodiani/gomontage/clip",
	)
	if err != nil {
		t.Fatalf("parsePackage() error: %v", err)
	}

	// Should find the Clip interface.
	var found bool
	for _, td := range pdoc.Types {
		if td.Name == "Clip" {
			found = true
			if !td.IsInterface {
				t.Error("Clip should be marked as interface")
			}
			if len(td.Fields) == 0 {
				t.Error("Clip interface should have methods listed as fields")
			}
			break
		}
	}
	if !found {
		t.Error("parsePackage should find Clip interface")
	}
}

func TestFormatDocComment(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect []string
	}{
		{
			name:   "simple",
			input:  "This is a doc comment.",
			expect: []string{"This is a doc comment."},
		},
		{
			name:   "with code block",
			input:  "Example:\n\n\tclip.NewVideo(\"test.mp4\")\n\tclip.Trim(0, 10)",
			expect: []string{"Example:", "```go", "clip.NewVideo(\"test.mp4\")", "```"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDocComment(tt.input)
			for _, s := range tt.expect {
				if !strings.Contains(result, s) {
					t.Errorf("formatDocComment() missing %q in:\n%s", s, result)
				}
			}
		})
	}
}

func TestFirstSentence(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"Simple doc.", "Simple doc"},
		{"First line.\nSecond line.", "First line"},
		{"", ""},
		{"No period", "No period"},
	}

	for _, tt := range tests {
		got := firstSentence(tt.input)
		if got != tt.expect {
			t.Errorf("firstSentence(%q) = %q, want %q", tt.input, got, tt.expect)
		}
	}
}

// --- Helpers ---

func readTemplate(t *testing.T, name string) string {
	t.Helper()
	data, err := templates.ReadFile("templates/" + name)
	if err != nil {
		t.Fatalf("could not read template %s: %v", name, err)
	}
	return string(data)
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read file %s: %v", path, err)
	}
	return string(data)
}
