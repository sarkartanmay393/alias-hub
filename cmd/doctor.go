package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sarkartanmay393/ah/pkg/manager"
	"github.com/spf13/cobra"
)

var doctorFix bool

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system health and dependencies",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running doctor...")

		// Auto-fix/Ensure environment is consistent
		if err := manager.EnsureDirs(); err != nil {
			fmt.Printf("[ERROR] Failed to ensure directories: %v\n", err)
			if !doctorFix {
				fmt.Println("  -> Hint: Try running 'ah doctor --fix' or 'ah init'")
			}
			return
		}

		// Check 1: Directory Structure (confirm what EnsureDirs created)
		root, err := manager.GetRootDir()
		if err != nil {
			fmt.Printf("[FAIL] Could not determine home directory: %v\n", err)
			return
		}
		fmt.Printf("[OK] Root directory exists at %s\n", root)

		// Check 2: Verify env.sh exists
		envPath := root + "/env.sh"
		if _, err := os.Stat(envPath); os.IsNotExist(err) {
			fmt.Printf("[WARN] env.sh not found at %s\n", envPath)
			if doctorFix {
				fmt.Println("  -> Regenerating env.sh...")
				if err := manager.GenerateEnvFile(); err != nil {
					fmt.Printf("  [FAIL] Failed to regenerate: %v\n", err)
				} else {
					fmt.Println("  [OK] Fixed.")
				}
			}
		} else {
			fmt.Println("[OK] env.sh exists.")
		}

		// Check 3: Dependencies
		if _, err := exec.LookPath("git"); err != nil {
			fmt.Println("[FAIL] 'git' is not installed or not in PATH.")
		} else {
			fmt.Println("[OK] 'git' is installed.")
		}

		// Check 4: Shell configuration
		home, _ := os.UserHomeDir()
		shell := os.Getenv("SHELL")
		var rcFile string
		if strings.Contains(shell, "zsh") {
			rcFile = home + "/.zshrc"
		} else {
			rcFile = home + "/.bashrc"
		}
		content, _ := os.ReadFile(rcFile)
		if strings.Contains(string(content), "AH_PATH") {
			fmt.Printf("[OK] Shell configured in %s\n", rcFile)
		} else {
			fmt.Printf("[WARN] Shell not configured. Run 'ah init' to set up.\n")
		}
	},
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorFix, "fix", false, "Attempt to fix found issues automatically")
	rootCmd.AddCommand(doctorCmd)
}
