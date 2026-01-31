package manager

import (
	"os"
	"path/filepath"
	"strings"
)

type ValidPackage struct {
	Name        string
	Description string
}

func SearchPackages(query string) ([]ValidPackage, error) {
	// Ensure directories exist first
	if err := EnsureDirs(); err != nil {
		return nil, err
	}

	var matches []ValidPackage

	// Use lock for thread-safe registry access
	err := WithLock(func() error {
		// Auto-update registry ensures we search fresh data
		if err := UpdateRegistry(); err != nil {
			return err
		}

		// ~/.ah/registry/registry
		baseDir, err := GetRegistryContentDir()
		if err != nil {
			return err
		}

		entries, err := os.ReadDir(baseDir)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		queryLower := strings.ToLower(query)

		for _, e := range entries {
			if !e.IsDir() {
				continue
			}

			// Load Metadata
			pkgPath := filepath.Join(baseDir, e.Name())
			meta, err := LoadMetadata(pkgPath)
			if err != nil {
				// Skip invalid packages (missing ah.yaml or too large)
				continue
			}

			nameMatch := strings.Contains(strings.ToLower(e.Name()), queryLower)
			descMatch := strings.Contains(strings.ToLower(meta.Description), queryLower)

			if nameMatch || descMatch {
				matches = append(matches, ValidPackage{Name: e.Name(), Description: meta.Description})
			}
		}
		return nil
	})

	return matches, err
}
