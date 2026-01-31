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

func RemovePackage(packageName string) error {
	return WithLock(func() error {
		root, err := GetRootDir()
		if err != nil {
			return err
		}

		// 1. Remove Symlink
		symlinkPath := filepath.Join(root, ActiveDir, packageName)
		if err := os.Remove(symlinkPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove active alias: %w", err)
		}

		// 2. Remove Source Repo
		// In Registry architecture, "Remove" only disables the package.
		// We process the removal from 'active' symlinks.
		// Logic to delete file from registry is permanently removed.
		if err := CompileAliases(); err != nil {
			fmt.Printf("Warning: Failed to compile aliases: %v\n", err)
		}
		return updateStateTimestamp()
	})
}
