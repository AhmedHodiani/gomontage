package main

import (
	"fmt"

	"github.com/ahmedhodiani/gomontage/project"
	"github.com/spf13/cobra"
)

func validateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate the current project structure",
		Long: `Checks that the current directory is a valid Gomontage project:
- gomontage.yaml exists and is valid
- main.go exists
- Required directories exist`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := project.LoadConfig("gomontage.yaml")
			if err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}

			fmt.Printf("Project:    %s\n", cfg.Name)
			fmt.Printf("Version:    %s\n", cfg.Version)
			fmt.Printf("Resolution: %dx%d\n", cfg.Resolution.Width, cfg.Resolution.Height)
			fmt.Printf("FPS:        %.0f\n", cfg.FPS)
			fmt.Printf("Output:     %s (%s)\n", cfg.Output.Directory, cfg.Output.Format)
			fmt.Println()
			fmt.Println("Project is valid.")

			return nil
		},
	}
}
