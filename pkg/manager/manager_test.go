package manager

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetRootDir(t *testing.T) {
	rootDir, err := GetRootDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should end with .ah
	if filepath.Base(rootDir) != RootDirName {
		t.Errorf("expected root dir to end with '%s', got '%s'", RootDirName, filepath.Base(rootDir))
	}

	// Should be absolute path
	if !filepath.IsAbs(rootDir) {
		t.Errorf("expected absolute path, got '%s'", rootDir)
	}
}

func TestEnsureDirs(t *testing.T) {
	// Create a temporary home directory for testing
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	err := EnsureDirs()
	if err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}

	// Check that directories were created
	rootDir := filepath.Join(tmpHome, RootDirName)
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		t.Errorf("root directory not created: %s", rootDir)
	}

	activeDir := filepath.Join(rootDir, ActiveDir)
	if _, err := os.Stat(activeDir); os.IsNotExist(err) {
		t.Errorf("active directory not created: %s", activeDir)
	}

	binDir := filepath.Join(rootDir, BinDir)
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		t.Errorf("bin directory not created: %s", binDir)
	}

	// Check env.sh was created
	envFile := filepath.Join(rootDir, EnvFile)
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		t.Errorf("env.sh not created: %s", envFile)
	}
}

func TestConflictError(t *testing.T) {
	conflicts := map[string]string{
		"ll": "package1",
		"gs": "package2",
	}
	err := &ConflictError{Conflicts: conflicts}

	msg := err.Error()
	if msg != "conflicts detected: 2 aliases collide" {
		t.Errorf("unexpected error message: %s", msg)
	}
}

func TestListPackages_EmptyActive(t *testing.T) {
	// Create a temporary home directory for testing
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Create directories manually
	rootDir := filepath.Join(tmpHome, RootDirName)
	activeDir := filepath.Join(rootDir, ActiveDir)
	os.MkdirAll(activeDir, 0755)

	packages, err := ListPackages()
	if err != nil {
		t.Fatalf("ListPackages failed: %v", err)
	}

	if len(packages) != 0 {
		t.Errorf("expected 0 packages, got %d", len(packages))
	}
}

func TestListPackages_WithPackages(t *testing.T) {
	// Create a temporary home directory for testing
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Create directories and fake packages
	rootDir := filepath.Join(tmpHome, RootDirName)
	activeDir := filepath.Join(rootDir, ActiveDir)
	os.MkdirAll(activeDir, 0755)

	// Create fake package directories
	os.MkdirAll(filepath.Join(activeDir, "package1"), 0755)
	os.MkdirAll(filepath.Join(activeDir, "package2"), 0755)
	// Hidden directory should be ignored
	os.MkdirAll(filepath.Join(activeDir, ".hidden"), 0755)

	packages, err := ListPackages()
	if err != nil {
		t.Fatalf("ListPackages failed: %v", err)
	}

	if len(packages) != 2 {
		t.Errorf("expected 2 packages, got %d: %v", len(packages), packages)
	}
}

func TestListPackages_NoActiveDir(t *testing.T) {
	// Create a temporary home directory for testing
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Don't create active directory
	packages, err := ListPackages()
	if err != nil {
		t.Fatalf("ListPackages failed: %v", err)
	}

	if len(packages) != 0 {
		t.Errorf("expected 0 packages for non-existent active dir, got %d", len(packages))
	}
}
