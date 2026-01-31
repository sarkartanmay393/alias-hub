package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/sarkartanmay393/ah/pkg/manager"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed alias packages",
	Run: func(cmd *cobra.Command, args []string) {
		showAll, _ := cmd.Flags().GetBool("all")

		// 1. Get Active (Installed) Packages
		activePkgs, err := manager.ListPackages()
		if err != nil {
			fmt.Printf("Error listing active packages: %v\n", err)
			return
		}

		activeMap := make(map[string]bool)
		for _, p := range activePkgs {
			activeMap[p] = true
		}

		// 2. Get Registry (Available) Packages - ONLY if requested
		var registryPkgs []string
		if showAll {
			var err error
			registryPkgs, err = manager.ListRegistryPackages()
			if err != nil {
				// If registry fails (e.g. not cloned yet), just show active
				registryPkgs = []string{}
			}
		}

		// Merge lists (unique)
		allMap := make(map[string]bool)
		for _, p := range activePkgs {
			allMap[p] = true
		}
		for _, p := range registryPkgs {
			allMap[p] = true
		}

		if len(allMap) == 0 {
			if showAll {
				fmt.Println("No packages found.")
			} else {
				fmt.Println("No installed packages.")
			}
			return
		}

		// Header
		fmt.Printf("%-20s %-12s %s\n", "PACKAGE", "STATUS", "DESCRIPTION")
		fmt.Println(algoLine(60)) // helper function needed or just hardcode line

		root, _ := manager.GetRootDir()

		for pkg := range allMap {
			status := "[Available]"
			if activeMap[pkg] {
				status = "[Enabled]"
			}

			// Try to get metadata from Registry info logic
			// If enabled, it's in active. If not, it's in registry.
			var metaPath string
			if activeMap[pkg] {
				metaPath = filepath.Join(root, manager.ActiveDir, pkg)
			} else {
				regPath, err := manager.GetRegistryPackagePath(pkg)
				if err == nil {
					metaPath = regPath
				}
			}

			desc := ""
			if metaPath != "" {
				meta, err := manager.LoadMetadata(metaPath)
				if err == nil {
					desc = meta.Description
				}
			}

			fmt.Printf("%-20s %-12s %s\n", pkg, status, desc)
		}
	},
}

func algoLine(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += "-"
	}
	return s
}

func init() {
	listCmd.Flags().BoolP("all", "a", false, "Show all available packages in registry")
	rootCmd.AddCommand(listCmd)
}
