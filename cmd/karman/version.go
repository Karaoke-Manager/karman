package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	versionCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "print version in JSON format")
	rootCmd.AddCommand(versionCmd)
}

// versionCmd is the command instance for the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long:  "Print information about the currently running version of Karman.",
	Args:  cobra.NoArgs,
	Run:   runVersion,
}

var (
	// The version string as set in the git tag.
	// The format is expected to be v1.2.3 or v1.2.3-pre where pre is a prerelease identifier.
	// Except for the "v" prefix the format should be a semantic version without build info.
	// Set via ldflags.
	version string
	// jsonOutput indicates whether the version should be printed in JSON format.
	jsonOutput bool
)

// runVersion actually prints the current Karman version.
func runVersion(_ *cobra.Command, _ []string) {
	info := getVersionInfo()
	if info == nil {
		_, _ = fmt.Fprintln(os.Stderr, "No version info provided during build.")
		if jsonOutput {
			fmt.Println("{}")
		}
		os.Exit(1)
	}
	if jsonOutput {
		e := json.NewEncoder(os.Stdout)
		e.SetIndent("", "  ")
		_ = e.Encode(info)
		return
	}

	if info.GitVersion != "" && info.GitTreeState != "dirty" {
		fmt.Printf("Karman Version %s (built with %s)\n", info.GitVersion, info.GoVersion)
	}
	if info.GitTreeState == "" {
		fmt.Printf("Karman Development Build (built with %s)\n", info.GoVersion)
		return
	}
	if len(info.GitCommit) > 12 {
		info.GitCommit = info.GitCommit[:12]
	}
	dirty := ""
	if info.GitTreeState == "dirty" {
		dirty = ".dirty"
	}
	fmt.Printf("Karman Version git.%s%s (built with %s)\n", info.GitCommit, dirty, info.GoVersion)
}

// versionInfo is the JSON schema for the version command output.
type versionInfo struct {
	Major        int    `json:"major"`
	Minor        int    `json:"minor"`
	Patch        int    `json:"patch"`
	Prerelease   string `json:"pre,omitempty"`
	GitVersion   string `json:"gitVersion"`
	GitCommit    string `json:"gitCommit"`
	GitTreeState string `json:"gitTreeState"`
	GoVersion    string `json:"goVersion"`
	Compiler     string `json:"compiler"`
	Platform     string `json:"platform"`
}

// getVersionInfo generates a versionInfo value from the version info provided at build time.
// If no build time version info was provided, nil is returned.
func getVersionInfo() *versionInfo {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return nil
	}
	v := &versionInfo{
		GoVersion: info.GoVersion,
	}
	if version != "" {
		v.GitVersion = version
		comps := strings.SplitN(strings.TrimPrefix(version, "v"), "-", 2)
		if len(comps) >= 2 {
			v.Prerelease = comps[1]
		}
		comps = strings.Split(comps[0], ".")
		if len(comps) >= 1 {
			v.Major, _ = strconv.Atoi(comps[0])
		}
		if len(comps) >= 2 {
			v.Minor, _ = strconv.Atoi(comps[1])
		}
		if len(comps) >= 3 {
			v.Patch, _ = strconv.Atoi(comps[2])
		}
	}

	var goos, goarch string
	fmt.Printf("%v", info.Settings)
	for _, setting := range info.Settings {
		switch setting.Key {
		case "-compiler":
			v.Compiler = setting.Value
		case "GOOS":
			goos = setting.Value
		case "GOARCH":
			goarch = setting.Value
		case "vcs.revision":
			v.GitCommit = setting.Value
		case "vcs.modified":
			if setting.Value == "true" {
				v.GitTreeState = "dirty"
			} else {
				v.GitTreeState = "clean"
			}
		}
	}
	v.Platform = goos + "/" + goarch
	return v
}
