package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"moreops/internal/core/config"
	"moreops/internal/validator"
)

var validateConfigPath string

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration and analyze Kubernetes manifests for security & consistency",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(validateConfigPath)
		if err != nil {
			return fmt.Errorf("configuration error: %w", err)
		}

		cmd.Printf("🔍 Validating configuration '%s'...\n", validateConfigPath)

		if cfg.Project.Name == "" {
			return fmt.Errorf("project name cannot be empty")
		}
		if len(cfg.Environments) == 0 {
			cmd.Println("⚠ Warning: No environments defined")
		}

		if cfg.Modules.Kubernetes != nil && cfg.Modules.Kubernetes.Enabled {
			if cfg.Modules.Kubernetes.Port < 1 || cfg.Modules.Kubernetes.Port > 65535 {
				return fmt.Errorf("invalid kubernetes port: %d (must be 1-65535)", cfg.Modules.Kubernetes.Port)
			}
			if cfg.Modules.Kubernetes.Replicas < 1 {
				return fmt.Errorf("replicas must be at least 1")
			}
		}

		cmd.Println("✓ Configuration syntax and parameters valid.")

		// Run Security and Consistency analysis on manifests
		val := validator.NewValidator()
		var allIssues []validator.Issue

		scanDirs := []string{"k8s", "charts"}
		for _, dir := range scanDirs {
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				cmd.Printf("🛡 Running security & consistency audit on %s/...\n", dir)
				issues, err := val.ValidateDirectory(dir)
				if err != nil {
					cmd.Printf("⚠ Error scanning %s: %v\n", dir, err)
				} else {
					allIssues = append(allIssues, issues...)
				}
			}
		}

		errorCount := 0
		warnCount := 0
		for _, issue := range allIssues {
			if issue.Severity == validator.SeverityError {
				errorCount++
				cmd.Printf("  ❌ [%s] %s: %s (%s)\n", issue.RuleID, issue.Resource, issue.Message, issue.File)
			} else {
				warnCount++
				cmd.Printf("  ⚠  [%s] %s: %s (%s)\n", issue.RuleID, issue.Resource, issue.Message, issue.File)
			}
		}

		if len(allIssues) == 0 {
			cmd.Println("✓ Manifest audit passed: No security or consistency issues found!")
		} else {
			cmd.Printf("\nAudit summary: %d error(s), %d warning(s)\n", errorCount, warnCount)
			if errorCount > 0 {
				return fmt.Errorf("manifest validation failed with %d error(s)", errorCount)
			}
		}

		return nil
	},
}

func init() {
	validateCmd.Flags().StringVarP(&validateConfigPath, "config", "c", config.DefaultConfigFileName, "Path to moreops.yaml")
	rootCmd.AddCommand(validateCmd)
}
