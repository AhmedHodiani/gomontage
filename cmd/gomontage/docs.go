package main

import (
	"fmt"

	"github.com/ahmedhodiani/gomontage/docs"
	"github.com/spf13/cobra"
)

func docsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Generate API documentation",
		Long: `Generates markdown documentation for the Gomontage API.

Guide pages (getting started, clips, timeline, effects, export profiles) are
copied from hand-written templates. API reference pages are auto-generated
from Go source code — every exported type, function, method, and constant
is documented with its full signature and doc comments.

The docs are written to a docs/ directory and include:
  - Hand-written guides (getting started, clips, timeline, effects, export)
  - Auto-generated per-package API reference (clip, timeline, effects, export)
  - API index with quick-reference tables`,
		RunE: func(cmd *cobra.Command, args []string) error {
			outputDir, _ := cmd.Flags().GetString("output")

			fmt.Printf("Generating documentation in %s/...\n", outputDir)

			if err := docs.Generate(outputDir); err != nil {
				return fmt.Errorf("doc generation failed: %w", err)
			}

			fmt.Println("Documentation generated successfully!")
			return nil
		},
	}

	cmd.Flags().StringP("output", "o", "docs", "Output directory for documentation")

	return cmd
}
