package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sarkartanmay393/ah/pkg/manager"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Completely remove Alias Hub and all data",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("⚠️  DANGER: This will delete:")
		fmt.Println("  - All installed alias packages")
		fmt.Println("  - The registry cache")
		fmt.Println("  - The entire ~/.ah directory")
		fmt.Println("  - Shell configuration lines in .zshrc/.bashrc")
		fmt.Println("")
		fmt.Print("Are you sure? Type 'DELETE' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)

		if response != "DELETE" {
			fmt.Println("Uninstall cancelled.")
			return
		}

		// 1. Remove Config from Shell
		removeShellConfig()

		// 2. Remove Data Directory
		root, _ := manager.GetRootDir()
		fmt.Printf("Removing %s...\n", root)
		if err := os.RemoveAll(root); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Println("✅ Uninstall complete.")

		// 3. Advise on Binary Removal
		execPath, err := os.Executable()
		if err == nil && strings.Contains(execPath, "Cellar") {
			fmt.Println("\n🍺 Homebrew installation detected.")
			fmt.Println("👉 To strictly remove the binary run: brew uninstall ah")
		} else {
			fmt.Println("\nTo remove the binary itself, delete:", execPath)
		}
	},
}

func removeShellConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	shell := os.Getenv("SHELL")
	var rcFile string
	if strings.Contains(shell, "zsh") {
		rcFile = filepath.Join(home, ".zshrc")
	} else {
		rcFile = filepath.Join(home, ".bashrc")
	}

	content, err := os.ReadFile(rcFile)
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	inBlock := false

	for _, line := range lines {
		// Check for block start marker (new style)
		if strings.Contains(line, "# >>> Alias Hub >>>") {
			inBlock = true
			continue
		}
		// Check for block end marker (new style)
		if strings.Contains(line, "# <<< Alias Hub <<<") {
			inBlock = false
			continue
		}
		// Legacy support: old style marker
		if strings.Contains(line, "# Alias Hub") && !inBlock {
			// Start of legacy block - skip until we find a non-matching line
			inBlock = true
			continue
		}
		if inBlock {
			// For legacy blocks, exit when we hit a line that's not part of our config
			if !strings.Contains(line, "AH_PATH") && !strings.Contains(line, "env.sh") && strings.TrimSpace(line) != "" {
				inBlock = false
				newLines = append(newLines, line)
			}
			continue
		}
		newLines = append(newLines, line)
	}

	// Write back only if changes were made
	if len(lines) != len(newLines) {
		if err := os.WriteFile(rcFile, []byte(strings.Join(newLines, "\n")), 0644); err == nil {
			fmt.Printf("Cleaned configuration from %s\n", rcFile)
		}
	}
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
