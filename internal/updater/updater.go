package updater

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Release represents a GitHub release
type Release struct {
	TagName    string    `json:"tag_name"`
	Name       string    `json:"name"`
	Draft      bool      `json:"draft"`
	Prerelease bool      `json:"prerelease"`
	CreatedAt  time.Time `json:"created_at"`
	Assets     []Asset   `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name        string `json:"name"`
	URL         string `json:"browser_download_url"`
	Size        int    `json:"size"`
	DownloadURL string `json:"url"`
}

// Config holds updater configuration
type Config struct {
	Owner          string // GitHub repo owner
	Repo           string // GitHub repo name
	CurrentVersion string // Current app version (e.g., "v1.2.3")
	BinaryName     string // Binary name to check for
	CheckInterval  time.Duration
	SkipPrerelease bool // Skip prerelease versions
	SkipDraft      bool // Skip draft versions
}

// Updater manages checking and applying updates
type Updater struct {
	config Config
	client *http.Client
}

// New creates a new updater instance
func New(config Config) *Updater {
	if config.CheckInterval == 0 {
		config.CheckInterval = 1 * time.Hour
	}
	if config.SkipPrerelease {
		config.SkipPrerelease = true
	}
	if config.SkipDraft {
		config.SkipDraft = true
	}

	return &Updater{
		config: config,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// CheckForUpdates checks if a new version is available
func (u *Updater) CheckForUpdates() (available bool, newVersion string, err error) {
	release, err := u.getLatestRelease()
	if err != nil {
		return false, "", err
	}

	if release == nil {
		return false, "", fmt.Errorf("no releases found")
	}

	// Compare versions
	if u.isNewerVersion(release.TagName, u.config.CurrentVersion) {
		return true, release.TagName, nil
	}

	return false, "", nil
}

// GetLatestRelease fetches the latest release information from GitHub
func (u *Updater) GetLatestRelease() (*Release, error) {
	return u.getLatestRelease()
}

// getLatestRelease fetches the latest release from GitHub API
func (u *Updater) getLatestRelease() (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", u.config.Owner, u.config.Repo)

	resp, err := u.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("repository not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	// Filter out draft/prerelease if configured
	if u.config.SkipDraft && release.Draft {
		return nil, fmt.Errorf("latest release is a draft")
	}
	if u.config.SkipPrerelease && release.Prerelease {
		return nil, fmt.Errorf("latest release is a prerelease")
	}

	return &release, nil
}

// GetReleaseInfo returns detailed information about a release
func (u *Updater) GetReleaseInfo(version string) (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", u.config.Owner, u.config.Repo, version)

	resp, err := u.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("release not found: %s", version)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	return &release, nil
}

// DownloadAsset downloads a release asset
func (u *Updater) DownloadAsset(asset Asset, dest string) error {
	resp, err := u.client.Get(asset.URL)
	if err != nil {
		return fmt.Errorf("failed to download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	file, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// FindAssetForPlatform finds the appropriate binary asset for the current platform
func (u *Updater) FindAssetForPlatform(release *Release) (*Asset, error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Map Go arch to GoReleaser arch naming (x86_64 instead of amd64, etc.)
	archMap := map[string]string{
		"amd64": "x86_64",
		"386":   "i386",
		"arm":   "armv7",
		"arm64": "arm64",
	}

	releaseArch, ok := archMap[arch]
	if !ok {
		releaseArch = arch // fallback to original arch name
	}

	// Build expected asset name patterns (handles both .exe and .zip formats)
	var expectedPatterns []string

	switch osName {
	case "linux":
		expectedPatterns = []string{
			fmt.Sprintf("%s_linux_%s", u.config.BinaryName, arch),
			fmt.Sprintf("%s_Linux_%s", u.config.BinaryName, arch),
			fmt.Sprintf("%s_linux_%s.tar.gz", u.config.BinaryName, arch),
			fmt.Sprintf("%s_Linux_%s.tar.gz", u.config.BinaryName, arch),
		}
	case "windows":
		expectedPatterns = []string{
			fmt.Sprintf("%s_windows_%s.exe", u.config.BinaryName, arch),
			fmt.Sprintf("%s_Windows_%s.exe", u.config.BinaryName, arch),
			fmt.Sprintf("%s_windows_%s.zip", u.config.BinaryName, releaseArch),
			fmt.Sprintf("%s_Windows_%s.zip", u.config.BinaryName, releaseArch),
		}
	case "darwin":
		expectedPatterns = []string{
			fmt.Sprintf("%s_darwin_%s", u.config.BinaryName, arch),
			fmt.Sprintf("%s_Darwin_%s", u.config.BinaryName, arch),
			fmt.Sprintf("%s_darwin_%s.tar.gz", u.config.BinaryName, arch),
			fmt.Sprintf("%s_Darwin_%s.tar.gz", u.config.BinaryName, arch),
		}
	default:
		return nil, fmt.Errorf("unsupported OS: %s", osName)
	}

	// Search for matching asset
	for _, asset := range release.Assets {
		for _, pattern := range expectedPatterns {
			if strings.Contains(asset.Name, pattern) {
				return &asset, nil
			}
		}
	}

	return nil, fmt.Errorf("no suitable asset found for %s/%s", osName, arch)
}

// isNewerVersion compares two version strings
// Simple semver comparison: v1.2.3 > v1.2.2
func (u *Updater) isNewerVersion(newVersion, currentVersion string) bool {
	// Remove 'v' prefix if present
	new := strings.TrimPrefix(strings.TrimSpace(newVersion), "v")
	current := strings.TrimPrefix(strings.TrimSpace(currentVersion), "v")

	// Simple string comparison (works for semver)
	// For production, consider using github.com/hashicorp/go-version
	return new > current
}

// ExtractAndApplyUpdate extracts the downloaded zip file and replaces the current binary
func (u *Updater) ExtractAndApplyUpdate(tmpFile string) error {
	// Open the zip file
	reader, err := zip.OpenReader(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer reader.Close()

	// Find the binary in the zip
	var binaryFile *zip.File
	binaryName := u.config.BinaryName
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	for _, file := range reader.File {
		if file.Name == binaryName {
			binaryFile = file
			break
		}
	}

	if binaryFile == nil {
		return fmt.Errorf("binary not found in zip: %s", binaryName)
	}

	// Extract binary to temporary location
	rc, err := binaryFile.Open()
	if err != nil {
		return fmt.Errorf("failed to open binary in zip: %w", err)
	}
	defer rc.Close()

	// Get the executable path
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Create backup of current binary
	backupPath := exePath + ".bak"
	if err := os.Rename(exePath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Write new binary
	outFile, err := os.Create(exePath)
	if err != nil {
		// Restore backup if extraction fails
		os.Rename(backupPath, exePath)
		return fmt.Errorf("failed to create new binary: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, rc); err != nil {
		outFile.Close()
		os.Remove(exePath)
		os.Rename(backupPath, exePath)
		return fmt.Errorf("failed to write new binary: %w", err)
	}
	outFile.Close()

	// Make binary executable on Unix
	if runtime.GOOS != "windows" {
		if err := os.Chmod(exePath, 0755); err != nil {
			return fmt.Errorf("failed to set executable permissions: %w", err)
		}
	}

	// Clean up temp file
	os.Remove(tmpFile)

	return nil
}

// ApplyUpdate runs the update command (for backwards compatibility)
func (u *Updater) ApplyUpdate(binaryPath string) error {
	cmd := exec.Command(binaryPath, "update")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("update command failed: %w", err)
	}

	return nil
}
