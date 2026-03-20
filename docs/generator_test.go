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

	expectedFiles := []string{
		"getting-started.md",
		"clips.md",
		"timeline.md",
		"cuts.md",
		"effects.md",
		"export.md",
		"api-reference.md",
	}

	for _, name := range expectedFiles {
		path := filepath.Join(outDir, name)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("expected file %s to exist: %v", name, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("expected file %s to be non-empty", name)
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

	// Write a dummy file that should be overwritten
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

func TestGettingStartedDocContent(t *testing.T) {
	content := gettingStartedDoc()
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
	content := clipsDoc()
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

func TestCutsDocContent(t *testing.T) {
	content := cutsDoc()
	cutTypes := []string{"Hard", "L-Cut", "J-Cut", "Dissolve", "CrossFade", "Jump Cut", "Dip to Black", "Wipe"}
	for _, ct := range cutTypes {
		if !strings.Contains(content, ct) {
			t.Errorf("cuts doc missing cut type: %s", ct)
		}
	}
}

func TestEffectsDocContent(t *testing.T) {
	content := effectsDoc()
	effects := []string{"FadeIn", "FadeOut", "SpeedUp", "Volume", "Normalize", "AudioSpeed"}
	for _, e := range effects {
		if !strings.Contains(content, e) {
			t.Errorf("effects doc missing effect: %s", e)
		}
	}
}

func TestExportDocContent(t *testing.T) {
	content := exportDoc()
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

func TestAPIReferenceDocContent(t *testing.T) {
	content := apiReferenceDoc()
	packages := []string{"clip", "timeline", "cuts", "effects", "export", "engine"}
	for _, pkg := range packages {
		if !strings.Contains(content, "`"+pkg+"`") {
			t.Errorf("api-reference doc missing package: %s", pkg)
		}
	}
}
