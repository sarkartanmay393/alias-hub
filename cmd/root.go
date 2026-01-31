package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sarkartanmay393/ah/pkg/manager"
	"github.com/sarkartanmay393/ah/pkg/updater"
	"github.com/sarkartanmay393/ah/pkg/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ah",
	Short: "Alias Hub - The ultimate shell alias manager",
	Long: `Alias Hub (ah) helps you manage, share, and sync shell aliases across your machines.
It features conflict detection, live updates, and a public registry.`,
	Version: version.Version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Background check for updates (non-blocking, with 24h debounce)
		go checkForUpdates()
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

const updateCheckInterval = 24 * time.Hour

func checkForUpdates() {
	// Get the check timestamp file path
	root, err := manager.GetRootDir()
	if err != nil {
		return
	}

	checkFile := filepath.Join(root, "last_update_check")

	// Check if we should skip (checked within last 24 hours)
	if info, err := os.Stat(checkFile); err == nil {
		if time.Since(info.ModTime()) < updateCheckInterval {
			return // Skip - checked recently
		}
	}

	// Update the check timestamp
	os.WriteFile(checkFile, []byte(time.Now().Format(time.RFC3339)), 0644)

	// Check for updates
	latestVersion, err := updater.CheckForUpdates()
	if err != nil {
		return // Silently fail
	}

	if latestVersion != "" {
		fmt.Fprintf(os.Stderr, "\n📦 Update available: v%s → v%s\n", version.Version, latestVersion)
		fmt.Fprintln(os.Stderr, "   Run 'ah self-update' to upgrade.")
	}
}

func Execute() error {
	return rootCmd.Execute()
}
