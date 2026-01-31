package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sarkartanmay393/ah/pkg/manager"
	"github.com/sarkartanmay393/ah/pkg/parser"
	"github.com/sarkartanmay393/ah/pkg/server"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install [package]",
	Short: "Install a package from the registry",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, pkgName := range args {
			fmt.Printf("\nInstalling %s...\n", pkgName)
			if err := manager.InstallPackage(pkgName); err != nil {
				// Check if it's a conflict error
				if conflictErr, ok := err.(*manager.ConflictError); ok {
					fmt.Println("\n[!] CONFLICTS DETECTED")
					fmt.Printf("Package '%s' has %d conflicting aliases.\n", pkgName, len(conflictErr.Conflicts))
					fmt.Print("Launch Web UI to resolve? [Y/n]: ")

					reader := bufio.NewReader(os.Stdin)
					response, _ := reader.ReadString('\n')
					response = strings.TrimSpace(strings.ToLower(response))

					if response == "" || response == "y" || response == "yes" {
						if err := server.Start(pkgName); err != nil {
							fmt.Printf("Error starting server: %v\n", err)
						}

						// Post-resolution summary
						fmt.Println("\nResolution session ended.")

						// 3. Compile (in case the server didn't, or to be safe)
						manager.CompileAliases()

						// 4. Show Status
						targetDir, _ := manager.GetRegistryPackagePath(pkgName)
						meta, _ := manager.LoadMetadata(targetDir)
						aliasPath := filepath.Join(targetDir, "alias.sh")
						aliases, _ := parser.ParseAliases(aliasPath)

						fmt.Printf("\n📦 Package: %s (%s)\n", meta.Name, meta.Version)
						fmt.Printf("✅ Installation Complete! %d aliases available.\n", len(aliases))
						continue
					}
					fmt.Println("Installation aborted.")
					continue
				}

				// Normal error
				fmt.Printf("Error installing package: %v\n", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
