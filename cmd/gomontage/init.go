package main

import (
	"fmt"

	"github.com/ahmedhodiani/gomontage/project"
	"github.com/spf13/cobra"
)

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init <project-name>",
		Short: "Create a new Gomontage project",
		Long: `Creates a new Gomontage project with the standard directory structure:

  <project-name>/
  ├── gomontage.yaml       Project configuration
  ├── main.go              Your editing script
  ├── go.mod               Go module file
  ├── .gitignore           Git ignore rules
  ├── resources/
  │   ├── video/           Video files (.mp4, .mov, etc.)
  │   ├── audio/           Audio files (.wav, .mp3, etc.)
  │   ├── images/          Images (.png, .jpg, etc.)
  │   └── fonts/           Custom fonts (.ttf, .otf)
  ├── output/              Rendered output
  └── temp/                Temporary files (auto-cleaned)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]

			fmt.Printf("Creating Gomontage project: %s\n", projectName)

			if err := project.Scaffold(projectName); err != nil {
				return fmt.Errorf("failed to create project: %w", err)
			}

			fmt.Println()
			fmt.Printf("Project created successfully!\n\n")
			fmt.Printf("Next steps:\n")
			fmt.Printf("  cd %s\n", projectName)
			fmt.Printf("  # Add your media files to resources/\n")
			fmt.Printf("  # Edit main.go to describe your video\n")
			fmt.Printf("  gomontage run\n")

			return nil
		},
	}
}
