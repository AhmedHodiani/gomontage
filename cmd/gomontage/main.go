package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version is set at build time via ldflags.
var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   "gomontage",
		Short: "Gomontage — programmatic video editing framework for Go",
		Long: `Gomontage is a Go framework for editing videos with code.

Write high-level Go code to trim clips, arrange them on tracks,
add transitions, layer audio, and export professional video.

Get started:
  gomontage init my-project    Create a new project
  gomontage run                Run the project's main.go
  gomontage probe video.mp4    Inspect a media file
  gomontage docs               Generate API documentation`,
		Version: version,
	}

	rootCmd.AddCommand(
		initCmd(),
		runCmd(),
		probeCmd(),
		validateCmd(),
		docsCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
