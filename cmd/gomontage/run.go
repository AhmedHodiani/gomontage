package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func runCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Run the current project's main.go",
		Long: `Executes the Gomontage project in the current directory by running
'go run main.go'. The project must contain a main.go file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check that main.go exists.
			if _, err := os.Stat("main.go"); os.IsNotExist(err) {
				return fmt.Errorf("no main.go found in current directory — are you in a Gomontage project?")
			}

			// Check that gomontage.yaml exists.
			if _, err := os.Stat("gomontage.yaml"); os.IsNotExist(err) {
				return fmt.Errorf("no gomontage.yaml found — are you in a Gomontage project?")
			}

			fmt.Println("Running Gomontage project...")
			fmt.Println()

			goRun := exec.Command("go", "run", "main.go")
			goRun.Stdout = os.Stdout
			goRun.Stderr = os.Stderr
			goRun.Stdin = os.Stdin

			if err := goRun.Run(); err != nil {
				return fmt.Errorf("project execution failed: %w", err)
			}

			return nil
		},
	}
}
