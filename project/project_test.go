package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig("test-project")
	if cfg.Name != "test-project" {
		t.Errorf("expected name 'test-project', got %q", cfg.Name)
	}
	if cfg.Resolution.Width != 1920 || cfg.Resolution.Height != 1080 {
		t.Errorf("expected 1920x1080, got %dx%d", cfg.Resolution.Width, cfg.Resolution.Height)
	}
	if cfg.FPS != 30 {
		t.Errorf("expected FPS 30, got %f", cfg.FPS)
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "gomontage.yaml")

	cfg := DefaultConfig("roundtrip-test")
	if err := cfg.Save(cfgPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loaded.Name != cfg.Name {
		t.Errorf("expected name %q, got %q", cfg.Name, loaded.Name)
	}
	if loaded.Resolution.Width != cfg.Resolution.Width {
		t.Errorf("expected width %d, got %d", cfg.Resolution.Width, loaded.Resolution.Width)
	}
	if loaded.FPS != cfg.FPS {
		t.Errorf("expected FPS %f, got %f", cfg.FPS, loaded.FPS)
	}
	if loaded.Temp.CleanOnExport != cfg.Temp.CleanOnExport {
		t.Errorf("expected clean_on_export %v, got %v", cfg.Temp.CleanOnExport, loaded.Temp.CleanOnExport)
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/gomontage.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestScaffold(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "my-video")

	if err := Scaffold(projectDir); err != nil {
		t.Fatalf("Scaffold failed: %v", err)
	}

	// Check expected files and directories exist.
	expectedDirs := []string{
		"resources/video",
		"resources/audio",
		"resources/images",
		"resources/fonts",
		"output",
		"temp",
	}
	for _, dir := range expectedDirs {
		path := filepath.Join(projectDir, dir)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("expected directory %q to exist: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected %q to be a directory", dir)
		}
	}

	expectedFiles := []string{
		"gomontage.yaml",
		"main.go",
		"go.mod",
		".gitignore",
	}
	for _, file := range expectedFiles {
		path := filepath.Join(projectDir, file)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("expected file %q to exist: %v", file, err)
			continue
		}
		if info.IsDir() {
			t.Errorf("expected %q to be a file, not directory", file)
		}
		if info.Size() == 0 {
			t.Errorf("expected file %q to not be empty", file)
		}
	}

	// Verify the config file is valid YAML.
	cfg, err := LoadConfig(filepath.Join(projectDir, "gomontage.yaml"))
	if err != nil {
		t.Fatalf("could not load scaffolded config: %v", err)
	}
	if cfg.Name != projectDir {
		t.Errorf("config name should match project dir, got %q", cfg.Name)
	}
}
