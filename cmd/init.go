package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sarkartanmay393/ah/pkg/manager"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize ah and setup shell configuration",
	Run: func(cmd *cobra.Command, args []string) {
		if err := manager.EnsureDirs(); err != nil {
			fmt.Printf("Error creating directories: %v\n", err)
			return
		}

		root, _ := manager.GetRootDir()
		configScript := fmt.Sprintf(`
# >>> Alias Hub >>>
export AH_PATH="%s"
[ -f "$AH_PATH/env.sh" ] && source "$AH_PATH/env.sh"
# <<< Alias Hub <<<
`, root)

		// Auto-Install Logic
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error: Could not find home directory.")
			return
		}

		shell := os.Getenv("SHELL")
		var rcFile string
		if strings.Contains(shell, "zsh") {
			rcFile = filepath.Join(home, ".zshrc")
		} else {
			rcFile = filepath.Join(home, ".bashrc")
		}

		// Check if file exists, create if not
		f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			fmt.Printf("Error opening %s: %v\n", rcFile, err)
			return
		}
		defer f.Close()

		// Read file to check for existence
		content, _ := os.ReadFile(rcFile)
		if strings.Contains(string(content), "export AH_PATH=") {
			fmt.Printf("✅ Alias Hub usage is already configured in %s\n", rcFile)
			return
		}

		// Append
		if _, err := f.WriteString(configScript); err != nil {
			fmt.Printf("Error writing to %s: %v\n", rcFile, err)
			return
		}

		fmt.Printf("✅ Setup complete! Added configuration to %s\n", rcFile)
		fmt.Println("👉 Please restart your terminal or run:")
		fmt.Printf("   source %s\n", rcFile)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
