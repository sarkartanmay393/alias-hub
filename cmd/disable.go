package cmd

import (
	"fmt"

	"github.com/sarkartanmay393/ah/pkg/manager"
	"github.com/spf13/cobra"
)

var disableCmd = &cobra.Command{
	Use:   "disable [package]",
	Short: "Disable an alias package (without removing it)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, packageName := range args {
			if err := manager.DisablePackage(packageName); err != nil {
				fmt.Printf("Error disabling package '%s': %v\n", packageName, err)
			} else {
				fmt.Printf("Package '%s' disabled.\n", packageName)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(disableCmd)
}
