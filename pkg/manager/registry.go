package manager

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	RegistryDir = "registry"
)

// UpdateRegistry ensures the registry is cloned and up to date
func UpdateRegistry() error {
	repoURL := os.Getenv("AH_REGISTRY_URL")
	if repoURL == "" {
		repoURL = RegistryRepo
	}

	root, err := GetRootDir()
	if err != nil {
		return err
	}
	registryPath := filepath.Join(root, RegistryDir)

	// Create a context with a 30-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		// Clone
		fmt.Printf("Cloning registry from %s...\n", repoURL)
		cmd := exec.CommandContext(ctx, "git", "clone", repoURL, registryPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Prevent interactive prompts
		cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "SSH_ASKPASS=/bin/false")

		if err := cmd.Run(); err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("clone timed out after 30s")
			}
			return fmt.Errorf("git clone failed: %w", err)
		}
		return nil
	}

	// Pull
	fmt.Println("Updating registry...")
	cmd := exec.CommandContext(ctx, "git", "-C", registryPath, "pull")
	cmd.Stderr = os.Stderr

	// Prevent interactive prompts
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "SSH_ASKPASS=/bin/false")

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Println("Warning: Registry update timed out (using cached data)")
		} else {
			fmt.Printf("Warning: Failed to update registry (using cached data): %v\n", err)
		}
		return nil // Soft fail: proceed with existing data
	}
	return nil
}

// GetRegistryContentDir returns the path where the actual packages are located (~/.ah/registry/registry)
func GetRegistryContentDir() (string, error) {
	root, err := GetRootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, RegistryDir, "registry"), nil
}

// GetRegistryPackagePath returns the absolute path to a package in the local registry
func GetRegistryPackagePath(packageName string) (string, error) {
	contentDir, err := GetRegistryContentDir()
	if err != nil {
		return "", err
	}

	pkgPath := filepath.Join(contentDir, packageName)

	if _, err := os.Stat(pkgPath); os.IsNotExist(err) {
		return "", fmt.Errorf("package '%s' not found in registry", packageName)
	}
	return pkgPath, nil
}

// ListRegistryPackages returns a list of all packages available in the local registry
func ListRegistryPackages() ([]string, error) {
	contentDir, err := GetRegistryContentDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(contentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var packages []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			packages = append(packages, entry.Name())
		}
	}
	return packages, nil
}
