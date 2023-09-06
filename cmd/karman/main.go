package main

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	cobra.EnableCommandSorting = false
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
