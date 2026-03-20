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

The docs are written to a docs/ directory and cover:
- Getting started guide
- Clip types and methods
- Timeline and track API
- Effects
- Export profiles`,
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
