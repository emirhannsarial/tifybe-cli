package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version is injected at build time via -ldflags. Defaults to "dev" for
// local builds so `tifybe --version` is always meaningful.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "tifybe",
	Short:   "Receive and inspect webhooks on localhost through a secure tunnel",
	Version: Version,
}

func Execute() {
	rootCmd.Version = Version
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
