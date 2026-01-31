package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ListPackages() ([]string, error) {
	root, err := GetRootDir()
	if err != nil {
		return nil, err
	}

	activePath := filepath.Join(root, ActiveDir)
	entries, err := os.ReadDir(activePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var packages []string
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), ".") {
			packages = append(packages, entry.Name())
		}
	}
	return packages, nil
}

// RemovePackage removes a package from the active directory.
// Returns an error if the package is not currently enabled.
func RemovePackage(packageName string) error {
	return WithLock(func() error {
		root, err := GetRootDir()
		if err != nil {
			return err
		}

		// 1. Check if package exists before removal
		symlinkPath := filepath.Join(root, ActiveDir, packageName)
		if _, err := os.Lstat(symlinkPath); os.IsNotExist(err) {
			return fmt.Errorf("package '%s' is not installed", packageName)
		}

		// 2. Remove Symlink
		if err := os.Remove(symlinkPath); err != nil {
			return fmt.Errorf("failed to remove package: %w", err)
		}

		// 3. Recompile aliases
		if err := CompileAliases(); err != nil {
			fmt.Printf("Warning: Failed to compile aliases: %v\n", err)
		}
		return updateStateTimestamp()
	})
}
