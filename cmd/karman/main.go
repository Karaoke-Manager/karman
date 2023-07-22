package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
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
