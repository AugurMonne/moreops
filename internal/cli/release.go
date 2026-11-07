package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"moreops/internal/core/config"
)

var releaseConfigPath string

var releaseCmd = &cobra.Command{
	Use:   "release [patch|minor|major]",
	Short: "Automate version bumping, changelog generation, and chart release",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bumpType := "patch"
		if len(args) > 0 {
			bumpType = strings.ToLower(args[0])
		}

		cfg, err := config.LoadConfig(releaseConfigPath)
		if err != nil {
			return err
		}

		oldVersion := cfg.Project.Version
		if oldVersion == "" {
			oldVersion = "0.1.0"
		}

		newVersion, err := bumpSemVer(oldVersion, bumpType)
		if err != nil {
			return err
		}

		cmd.Printf("Bumping version: %s -> %s (%s)\n", oldVersion, newVersion, bumpType)

		// 1. Update moreops.yaml
		cfg.Project.Version = newVersion
		if err := config.SaveConfig(releaseConfigPath, cfg); err != nil {
			return fmt.Errorf("failed updating moreops.yaml: %w", err)
		}
		cmd.Println("✓ Updated", releaseConfigPath)

		// 2. Update Helm Chart.yaml if present
		chartPath := filepath.Join("charts", cfg.Project.Name, "Chart.yaml")
		if _, err := os.Stat(chartPath); err == nil {
			if err := updateChartYaml(chartPath, newVersion); err != nil {
				cmd.Printf("⚠ Warning: failed updating Chart.yaml: %v\n", err)
			} else {
				cmd.Printf("✓ Updated %s to version %s\n", chartPath, newVersion)
			}
		}

		// 3. Update CHANGELOG.md
		changelogPath := "CHANGELOG.md"
		if err := appendChangelog(changelogPath, newVersion); err != nil {
			cmd.Printf("⚠ Warning: failed writing changelog: %v\n", err)
		} else {
			cmd.Println("✓ Updated CHANGELOG.md")
		}

		cmd.Println("\n==========================================")
		cmd.Printf("✓ Release %s prepared successfully!\n", newVersion)
		cmd.Println("==========================================")
		cmd.Printf("To publish this release to git, run:\n")
		cmd.Printf("  git commit -am \"chore: release v%s\"\n", newVersion)
		cmd.Printf("  git tag -a v%s -m \"Release v%s\"\n", newVersion, newVersion)
		cmd.Printf("  git push origin v%s\n", newVersion)

		return nil
	},
}

func bumpSemVer(current, bumpType string) (string, error) {
	current = strings.TrimPrefix(current, "v")
	parts := strings.Split(current, ".")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid semver: %s (expected x.y.z)", current)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", err
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", err
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", err
	}

	switch bumpType {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	default:
		return "", fmt.Errorf("unknown bump type '%s': choose patch, minor, or major", bumpType)
	}

	return fmt.Sprintf("%d.%d.%d", major, minor, patch), nil
}

func updateChartYaml(path, newVersion string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return err
	}

	// Regex replace version and appVersion preserving comments
	vRegex := regexp.MustCompile(`(?m)^version:\s*.*$`)
	appVRegex := regexp.MustCompile(`(?m)^appVersion:\s*.*$`)

	updated := vRegex.ReplaceAllString(string(data), fmt.Sprintf("version: %s", newVersion))
	updated = appVRegex.ReplaceAllString(updated, fmt.Sprintf("appVersion: %q", newVersion))

	return os.WriteFile(path, []byte(updated), 0644)
}

func appendChangelog(path, version string) error {
	dateStr := time.Now().Format("2006-01-02")
	entry := fmt.Sprintf("\n## [%s] - %s\n- Automated release by MoreOps\n- Infrastructure and manifest configurations updated\n",
		version, dateStr)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		header := "# Changelog\nAll notable changes to this project will be documented in this file.\n"
		return os.WriteFile(path, []byte(header+entry), 0644)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(entry)
	return err
}

func init() {
	releaseCmd.Flags().StringVarP(&releaseConfigPath, "config", "c", config.DefaultConfigFileName, "Path to moreops.yaml")
	rootCmd.AddCommand(releaseCmd)
}
