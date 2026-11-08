package cli

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	upgradeCheckOnly bool
	upgradeForce     bool
	upgradeRepo      = "AugurMonne/moreops"
)

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Name    string        `json:"name"`
	Body    string        `json:"body"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

var upgradeCmd = &cobra.Command{
	Use:     "upgrade",
	Aliases: []string{"update", "self-update"},
	Short:   "Upgrade MoreOps CLI to the latest version",
	Long: `Checks GitHub for the latest release of MoreOps, downloads the appropriate
pre-compiled binary for your operating system and architecture, and seamlessly
replaces the current executable in-place.`,
	Example: `  # Check for updates and upgrade to latest release
  moreops upgrade

  # Check if a new version is available without downloading
  moreops upgrade --check

  # Force reinstall current version
  moreops upgrade --force`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpgrade(cmd.OutOrStdout(), upgradeCheckOnly, upgradeForce)
	},
}

func init() {
	upgradeCmd.Flags().BoolVarP(&upgradeCheckOnly, "check", "c", false, "Only check if an update is available without applying it")
	upgradeCmd.Flags().BoolVarP(&upgradeForce, "force", "f", false, "Force reinstall even if current version is up to date")
	rootCmd.AddCommand(upgradeCmd)
}

func runUpgrade(out io.Writer, checkOnly, force bool) error {
	fmt.Fprintf(out, "Checking for latest MoreOps release from github.com/%s...\n", upgradeRepo)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	release, err := fetchLatestRelease(ctx, upgradeRepo)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	currentVer := strings.TrimPrefix(Version, "v")
	latestVer := strings.TrimPrefix(release.TagName, "v")

	isNewer := compareVersions(latestVer, currentVer) > 0

	if !isNewer && !force {
		fmt.Fprintf(out, "✓ MoreOps is already up to date (version %s).\n", Version)
		return nil
	}

	if checkOnly {
		if isNewer {
			fmt.Fprintf(out, "✦ A newer version is available: %s -> %s\nRun 'moreops upgrade' to install.\n", Version, release.TagName)
		} else {
			fmt.Fprintf(out, "✓ MoreOps is up to date (version %s).\n", Version)
		}
		return nil
	}

	fmt.Fprintf(out, "Upgrading MoreOps from %s to %s...\n", Version, release.TagName)

	// Determine matching asset name: moreops_<os>_<arch>.tar.gz
	osName := runtime.GOOS
	archName := runtime.GOARCH
	expectedAsset := fmt.Sprintf("moreops_%s_%s.tar.gz", osName, archName)

	var targetAsset *githubAsset
	for _, a := range release.Assets {
		if a.Name == expectedAsset {
			targetAsset = &a
			break
		}
	}

	if targetAsset == nil {
		return fmt.Errorf("no release binary found for your platform (%s/%s). Expected asset '%s'", osName, archName, expectedAsset)
	}

	// Find the current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine current executable path: %w", err)
	}
	realPath, err := filepath.EvalSymlinks(execPath)
	if err == nil {
		execPath = realPath
	}

	fmt.Fprintf(out, "Downloading %s (%.2f MB)...\n", targetAsset.Name, float64(targetAsset.Size)/(1024*1024))

	downloadReq, err := http.NewRequestWithContext(ctx, "GET", targetAsset.BrowserDownloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}
	downloadReq.Header.Set("User-Agent", "moreops-updater")

	resp, err := http.DefaultClient.Do(downloadReq)
	if err != nil {
		return fmt.Errorf("failed to download release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with HTTP %d %s", resp.StatusCode, resp.Status)
	}

	// Extract binary from tar.gz in memory
	binaryData, err := extractBinaryFromTarGz(resp.Body, "moreops")
	if err != nil {
		return fmt.Errorf("failed to extract binary from archive: %w", err)
	}

	// Prepare temporary replacement file in the same directory (to preserve filesystem boundary for atomic rename)
	execDir := filepath.Dir(execPath)
	tmpFile, err := os.CreateTemp(execDir, "moreops-upgrade-*")
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied writing to %s. Please run with sudo:\n  sudo moreops upgrade", execDir)
		}
		return fmt.Errorf("failed to create temp file in %s: %w", execDir, err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(binaryData); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write upgraded binary: %w", err)
	}
	if err := tmpFile.Chmod(0755); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to make binary executable: %w", err)
	}
	tmpFile.Close()

	// Atomic replace
	if err := os.Rename(tmpPath, execPath); err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied replacing %s. Please run with sudo:\n  sudo moreops upgrade", execPath)
		}
		return fmt.Errorf("failed to replace binary at %s: %w", execPath, err)
	}

	fmt.Fprintf(out, "\n✓ Successfully upgraded MoreOps to %s!\nBinary updated at: %s\n", release.TagName, execPath)
	return nil
}

func fetchLatestRelease(ctx context.Context, repo string) (*githubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "moreops-updater")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no published releases found on https://github.com/%s", repo)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("failed to parse release metadata: %w", err)
	}
	return &rel, nil
}

func extractBinaryFromTarGz(r io.Reader, binaryName string) ([]byte, error) {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		baseName := filepath.Base(header.Name)
		if baseName == binaryName || baseName == binaryName+".exe" {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil, err
			}
			return data, nil
		}
	}

	return nil, fmt.Errorf("binary '%s' not found inside archive", binaryName)
}

// compareVersions returns 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(strings.TrimSpace(v1), ".")
	parts2 := strings.Split(strings.TrimSpace(v2), ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &n1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &n2)
		}
		if n1 > n2 {
			return 1
		}
		if n1 < n2 {
			return -1
		}
	}
	return 0
}
