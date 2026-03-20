package project

import (
	"fmt"
	"os"
	"path/filepath"

	gomontage "github.com/ahmedhodiani/gomontage"
)

// Scaffold creates a new Gomontage project with the standard directory structure
// and boilerplate files.
func Scaffold(projectName string) error {
	// Create the project root directory.
	if err := os.MkdirAll(projectName, 0755); err != nil {
		return fmt.Errorf("could not create project directory: %w", err)
	}

	// Create subdirectories.
	dirs := []string{
		"resources/video",
		"resources/audio",
		"resources/images",
		"resources/fonts",
		"output",
		"temp",
	}
	for _, dir := range dirs {
		path := filepath.Join(projectName, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("could not create directory %q: %w", dir, err)
		}
	}

	// Create gomontage.yaml config.
	cfg := DefaultConfig(projectName)
	cfgPath := filepath.Join(projectName, "gomontage.yaml")
	if err := cfg.Save(cfgPath); err != nil {
		return err
	}

	// Create main.go boilerplate.
	mainPath := filepath.Join(projectName, "main.go")
	if err := os.WriteFile(mainPath, []byte(mainGoTemplate(projectName)), 0644); err != nil {
		return fmt.Errorf("could not write main.go: %w", err)
	}

	// Create go.mod.
	goModPath := filepath.Join(projectName, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModTemplate(projectName)), 0644); err != nil {
		return fmt.Errorf("could not write go.mod: %w", err)
	}

	// Create .gitignore.
	gitignorePath := filepath.Join(projectName, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignoreTemplate()), 0644); err != nil {
		return fmt.Errorf("could not write .gitignore: %w", err)
	}

	return nil
}

func mainGoTemplate(projectName string) string {
	return `package main

import (
	"fmt"
	"time"

	"github.com/ahmedhodiani/gomontage/clip"
	"github.com/ahmedhodiani/gomontage/export"
	"github.com/ahmedhodiani/gomontage/timeline"
)

func main() {
	// Create a new timeline at 1080p, 30fps.
	tl := timeline.New(timeline.Config{
		Width:  1920,
		Height: 1080,
		FPS:    30,
	})

	// ---------------------------------------------------------------
	// Load your media files from the resources/ directory.
	// ---------------------------------------------------------------

	// Example: Load a video clip and trim it.
	// video := clip.NewVideo("resources/video/footage.mp4")
	// intro := video.Trim(0, 10*time.Second)

	// Example: Load audio files.
	// narration := clip.NewAudio("resources/audio/narration.wav")
	// music := clip.NewAudio("resources/audio/music.mp3").WithVolume(0.3)

	// ---------------------------------------------------------------
	// Create tracks and arrange your clips.
	// ---------------------------------------------------------------

	// Video track for the main visuals.
	mainVideo := tl.AddVideoTrack("main")

	// Audio track for voiceover/narration.
	// voiceover := tl.AddAudioTrack("narration")

	// Audio track for background music.
	// bgMusic := tl.AddAudioTrack("music")

	// ---------------------------------------------------------------
	// Place clips on tracks.
	// ---------------------------------------------------------------

	// Example: Place clips sequentially on the video track.
	// mainVideo.AddSequence(intro, middleSection, outro)

	// Example: Place audio at specific times.
	// voiceover.Add(narration, timeline.At(0))
	// bgMusic.Add(music.WithFadeIn(2*time.Second).WithFadeOut(3*time.Second), timeline.At(0))

	// ---------------------------------------------------------------
	// Export the final video.
	// ---------------------------------------------------------------

	_ = mainVideo   // Remove this line once you add real clips.
	_ = time.Second // Remove this line once you use time.
	_ = clip.TypeVideo  // Remove this line once you use clip.
	_ = export.YouTube1080p // Remove this line once you use export.

	// Uncomment to export:
	// tl.Export(export.YouTube1080p(), "output/final.mp4")

	fmt.Printf("` + projectName + ` - Gomontage project ready!\n")
	fmt.Println("Edit this file to start building your video.")
	fmt.Println("See: https://github.com/ahmedhodiani/gomontage for documentation.")
}
`
}

func goModTemplate(projectName string) string {
	return fmt.Sprintf(`module %s

go %s

require github.com/ahmedhodiani/gomontage v%s
`, projectName, gomontage.GoVersion, gomontage.Version)
}

func gitignoreTemplate() string {
	return `# Output
output/

# Temp files
temp/

# Binary
*.exe
*.exe~
*.dll
*.so
*.dylib

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db
`
}
