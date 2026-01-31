package cmd

import (
	"fmt"

	"github.com/sarkartanmay393/ah/pkg/manager"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the package registry and re-compile aliases",
	Long:  `Downloads the latest package definitions from the registry and re-generates your alias configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := manager.EnsureDirs(); err != nil {
			fmt.Printf("Error ensuring directories: %v\n", err)
			return
		}

		// Use WithLock for thread-safe registry update and compile
		if err := manager.WithLock(func() error {
			fmt.Println("Updating registry...")
			if err := manager.UpdateRegistry(); err != nil {
				return fmt.Errorf("registry update failed: %w", err)
			}

			fmt.Println("Compiling aliases...")
			if err := manager.CompileAliases(); err != nil {
				return fmt.Errorf("compile failed: %w", err)
			}
			return nil
		}); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Println("All set! Registry and aliases updated.")
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
