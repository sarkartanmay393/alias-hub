// Package updater provides self-update functionality for the ah CLI.
// It checks for new releases on GitHub and downloads/installs updates.
package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/sarkartanmay393/ah/pkg/version"
)

// GitHub repository coordinates for release checks.
const (
	RepoOwner = "sarkartanmay393"
	RepoName  = "ah"
)

// Release represents a GitHub release response.
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a downloadable file attached to a release.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// CheckForUpdates returns the latest version tag if it's newer than current
func CheckForUpdates() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", RepoOwner, RepoName)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch latest release: %s", resp.Status)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(version.Version, "v")

	if isNewerVersion(latest, current) {
		return latest, nil
	}
	return "", nil
}

// isNewerVersion compares two semver strings and returns true if latest > current
func isNewerVersion(latest, current string) bool {
	latestParts := strings.Split(latest, ".")
	currentParts := strings.Split(current, ".")

	// Pad to ensure same length
	for len(latestParts) < 3 {
		latestParts = append(latestParts, "0")
	}
	for len(currentParts) < 3 {
		currentParts = append(currentParts, "0")
	}

	for i := 0; i < 3; i++ {
		l := parseVersionPart(latestParts[i])
		c := parseVersionPart(currentParts[i])
		if l > c {
			return true
		}
		if l < c {
			return false
		}
	}
	return false // Equal
}

// parseVersionPart extracts the numeric part of a version segment
func parseVersionPart(s string) int {
	// Handle pre-release suffixes like "1-beta"
	s = strings.Split(s, "-")[0]
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

// SelfUpdate downloads and installs the latest version
func SelfUpdate() error {
	latestVersion, err := CheckForUpdates()
	if err != nil {
		return fmt.Errorf("check failed: %w", err)
	}
	if latestVersion == "" {
		fmt.Println("Already on the latest version.")
		return nil
	}

	fmt.Printf("Upgrading from %s to %s...\n", version.Version, latestVersion)

	// Determine asset name
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Map typical Go arch to release naming convention if needed
	// Assuming install.sh naming: ah-<os>-<arch>
	assetName := fmt.Sprintf("ah-%s-%s", osName, arch)

	url := fmt.Sprintf("https://github.com/%s/%s/releases/download/v%s/%s", RepoOwner, RepoName, latestVersion, assetName)

	// Download with timeout
	tmpFile, err := os.CreateTemp("", "ah-update")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()

	// Use client with timeout
	client := &http.Client{Timeout: 60 * time.Second}

	fmt.Printf("Downloading %s...\n", url)
	resp, err := client.Get(url)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return err
	}

	// Close file before rename (required on some systems)
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	// Make executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return err
	}

	// Move to current executable path
	executablePath, err := os.Executable()
	if err != nil {
		os.Remove(tmpPath)
		return err
	}

	// On Linux/Mac we can rename over running executable usually
	if err := os.Rename(tmpPath, executablePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to replace binary (try running with sudo): %w", err)
	}

	fmt.Println("Update successful!")
	return nil
}
