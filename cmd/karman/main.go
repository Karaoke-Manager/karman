package main

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	cobra.EnableCommandSorting = false
}

// main is the entrypoint of the karman application.
func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
