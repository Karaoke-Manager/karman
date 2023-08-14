package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "karman",
	Short: "Karman - The Karaoke Manager",
	Long:  `The Karaoke Manager helps you organize your UltraStar Karaoke songs.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
