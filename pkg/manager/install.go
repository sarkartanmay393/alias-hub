package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bufio"

	"github.com/sarkartanmay393/ah/pkg/parser"
)

// InstallPackage installs a package from the central registry
func InstallPackage(packageName string) error {
	if err := EnsureDirs(); err != nil {
		return err
	}

	// Phase 1: Update registry and validate package (with lock)
	var meta *PackageMetadata
	var aliases []parser.AliasDef
	var targetDir string

	err := WithLock(func() error {
		// 1. Update Registry
		if err := UpdateRegistry(); err != nil {
			return fmt.Errorf("failed to update registry: %w", err)
		}

		// 2. Find Package
		var err error
		targetDir, err = GetRegistryPackagePath(packageName)
		if err != nil {
			return err
		}

		// 3. Validate Package Structure & Load Metadata
		meta, err = LoadMetadata(targetDir)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("invalid package: 'ah.yaml' missing in %s", packageName)
			}
			return fmt.Errorf("invalid package metadata: %w", err)
		}

		aliasPath := filepath.Join(targetDir, "alias.sh")
		if _, err := os.Stat(aliasPath); os.IsNotExist(err) {
			return fmt.Errorf("invalid package: 'alias.sh' missing in %s", packageName)
		}

		// 4. Conflict Check (ATOMIC due to lock)
		conflicts, err := CheckConflicts(targetDir)
		if err != nil {
			fmt.Printf("Warning: Failed to check conflicts: %v\n", err)
		}
		if len(conflicts) > 0 {
			return &ConflictError{Conflicts: conflicts}
		}

		// 5. Parse aliases for preview
		aliases, _ = parser.ParseAliases(aliasPath)

		return nil
	})

	if err != nil {
		return err
	}

	// Phase 2: Show preview and prompt user (NO LOCK - avoids starvation)
	fmt.Printf("\n📦 Package: %s (%s)\n", meta.Name, meta.Version)
	fmt.Printf("📝 Desc:    %s\n", meta.Description)
	fmt.Printf("👤 Author:  %s\n", meta.Author)
	if meta.Website != "" {
		fmt.Printf("🔗 Web:     %s\n", meta.Website)
	}
	fmt.Printf("\nContains %d aliases:\n", len(aliases))
	for _, a := range aliases {
		fmt.Printf("  %s = %s\n", a.Name, a.Command)
	}
	fmt.Print("\nProceed to enable? [Y/n]: ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "" && response != "y" && response != "yes" {
		fmt.Println("Package installed but NOT enabled. Use 'ah enable' later.")
		return nil
	}

	// Phase 3: Enable package (with lock again)
	return EnablePackage(packageName)
}

// EnablePackage links a package from the REGISTRY to active
func EnablePackage(packageName string) error {
	return WithLock(func() error {
		return enablePackageInternal(packageName)
	})
}

// enablePackageInternal performs the symlink and compile updates.
// Assumes LOCK IS HELD.
func enablePackageInternal(packageName string) error {
	root, err := GetRootDir()
	if err != nil {
		return err
	}

	// Source is now in REGISTRY (Monorepo structure: registry/pkg)
	contentDir, err := GetRegistryContentDir()
	if err != nil {
		return err
	}
	source := filepath.Join(contentDir, packageName)
	target := filepath.Join(root, ActiveDir, packageName)

	if _, err := os.Stat(source); os.IsNotExist(err) {
		return fmt.Errorf("package %s not found in local registry", packageName)
	}

	// Remove existing symlink if any (re-enable)
	os.Remove(target)

	if err := os.Symlink(source, target); err != nil {
		return fmt.Errorf("failed to symlink: %w", err)
	}

	fmt.Printf("Enabled package: %s\n", packageName)

	// Internal update
	if err := CompileAliases(); err != nil {
		fmt.Printf("Warning: Failed to compile aliases: %v\n", err)
	}
	return updateStateTimestamp()
}
