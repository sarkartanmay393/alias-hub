package manager

import (
	"fmt"
	"os"
	"path/filepath"
)

// DisablePackage removes the symlink from the active directory
func DisablePackage(packageName string) error {
	return WithLock(func() error {
		root, err := GetRootDir()
		if err != nil {
			return err
		}

		symlinkPath := filepath.Join(root, ActiveDir, packageName)

		// Check if enabled
		if _, err := os.Lstat(symlinkPath); os.IsNotExist(err) {
			return fmt.Errorf("package %s is not enabled", packageName)
		}

		if err := os.Remove(symlinkPath); err != nil {
			return fmt.Errorf("failed to disable package: %w", err)
		}

		fmt.Printf("Disabled package: %s\n", packageName)

		if err := CompileAliases(); err != nil {
			fmt.Printf("Warning: Failed to compile aliases: %v\n", err)
		}
		return updateStateTimestamp()
	})
}

// EnablePackageFromRepo enables an already installed package from the registry.
// Returns an error if the package is already enabled or not installed.
func EnablePackageFromRepo(packageName string) error {
	root, err := GetRootDir()
	if err != nil {
		return err
	}

	// Check if already enabled
	symlinkPath := filepath.Join(root, ActiveDir, packageName)
	if _, err := os.Lstat(symlinkPath); err == nil {
		return fmt.Errorf("package '%s' is already enabled", packageName)
	}

	// Verify package exists in registry
	contentDir, err := GetRegistryContentDir()
	if err != nil {
		return err
	}
	repoPath := filepath.Join(contentDir, packageName)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return fmt.Errorf("package %s is not installed (use 'ah install')", packageName)
	}

	// Use the existing EnablePackage function from install.go
	return EnablePackage(packageName)
}
