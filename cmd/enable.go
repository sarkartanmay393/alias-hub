package cmd

import (
	"fmt"

	"github.com/sarkartanmay393/ah/pkg/manager"
	"github.com/spf13/cobra"
)

var enableCmd = &cobra.Command{
	Use:   "enable [package]",
	Short: "Enable an installed alias package",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, packageName := range args {
			if err := manager.EnablePackageFromRepo(packageName); err != nil {
				fmt.Printf("Error enabling package '%s': %v\n", packageName, err)
			} else {
				fmt.Printf("Package '%s' enabled.\n", packageName)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(enableCmd)
}
