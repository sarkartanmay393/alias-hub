package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/sarkartanmay393/ah/pkg/version"
)

const (
	RepoOwner = "sarkartanmay393"
	RepoName  = "ah" // Updated repo name
)

type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// CheckForUpdates returns the latest version tag if it's newer than current
func CheckForUpdates() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", RepoOwner, RepoName)
	resp, err := http.Get(url)
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

	if latest != current {
		return latest, nil
	}
	return "", nil
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

	// Download
	tmpFile, err := os.CreateTemp("", "ah-update")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	fmt.Printf("Downloading %s...\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return err
	}
	// Make executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return err
	}

	// Move to current executable path
	executablePath, err := os.Executable()
	if err != nil {
		return err
	}

	// Check write permissions sort of by trying to move (sudo check)
	// We can't elevate privileges from here easily, user must run as sudo if needed.

	// On Linux/Mac we can rename over running executable usually?
	// safest is usually rename old aside, move new in.

	if err := os.Rename(tmpFile.Name(), executablePath); err != nil {
		return fmt.Errorf("failed to replace binary (try running with sudo): %w", err)
	}

	fmt.Println("Update successful!")
	return nil
}
