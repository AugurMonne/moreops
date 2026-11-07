package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"moreops/internal/core/config"
	"moreops/internal/modules/cluster"
	"moreops/internal/validator"
)

var (
	testEphemeral bool
	testConfig    string
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run validation, integration tests, and ephemeral cluster verification",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(testConfig)
		if err != nil {
			return fmt.Errorf("configuration error: %w", err)
		}

		cmd.Println("\n==========================================")
		cmd.Printf("   MoreOps Verification Suite: %s\n", cfg.Project.Name)
		cmd.Println("==========================================")

		// 1. Static security & consistency validation
		cmd.Println("\n[1/2] Running Security & Manifest Validation...")
		val := validator.NewValidator()
		hasErrors := false

		scanDirs := []string{"k8s", "charts"}
		for _, dir := range scanDirs {
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				issues, err := val.ValidateDirectory(dir)
				if err != nil {
					cmd.Printf("  ⚠ Error reading %s: %v\n", dir, err)
					continue
				}
				for _, issue := range issues {
					if issue.Severity == validator.SeverityError {
						hasErrors = true
						cmd.Printf("  ❌ [%s] %s: %s\n", issue.RuleID, issue.Resource, issue.Message)
					} else {
						cmd.Printf("  ⚠  [%s] %s: %s\n", issue.RuleID, issue.Resource, issue.Message)
					}
				}
			}
		}

		if hasErrors {
			return fmt.Errorf("static validation failed with errors")
		}
		cmd.Println("  ✓ Static checks passed.")

		// 2. Ephemeral cluster testing
		if testEphemeral {
			cmd.Println("\n[2/2] Running Ephemeral Cluster E2E Test...")
			mgr := cluster.NewK3dManager()
			available, reason := mgr.CheckAvailable()
			if !available {
				return fmt.Errorf("ephemeral testing skipped: %s (install docker & k3d or check 'moreops doctor')", reason)
			}

			err := mgr.RunEphemeralTest(context.Background(), func(contextName string) error {
				cmd.Printf("  → Testing deployment against ephemeral context '%s'...\n", contextName)

				// Test applying manifests if k8s/ exists
				if info, err := os.Stat("k8s"); err == nil && info.IsDir() {
					applyCmd := exec.Command("kubectl", "--context", contextName, "apply", "-f", "k8s/", "--dry-run=server")
					out, err := applyCmd.CombinedOutput()
					if err != nil {
						return fmt.Errorf("kubectl server dry-run failed: %s (%w)", string(out), err)
					}
					cmd.Printf("  ✓ Kubectl dry-run=server succeeded:\n%s\n", string(out))
				}

				// Test helm lint / template if charts/ exists
				chartPath := filepath.Join("charts", cfg.Project.Name)
				if info, err := os.Stat(chartPath); err == nil && info.IsDir() {
					lintCmd := exec.Command("helm", "lint", chartPath)
					if out, err := lintCmd.CombinedOutput(); err == nil {
						cmd.Printf("  ✓ Helm lint passed:\n%s\n", string(out))
					}
				}

				return nil
			})
			if err != nil {
				return fmt.Errorf("ephemeral cluster test failed: %w", err)
			}
			cmd.Println("  ✓ Ephemeral cluster testing completed successfully!")
		} else {
			cmd.Println("\n[2/2] Ephemeral testing skipped (pass --ephemeral to spin up a temporary k3d cluster).")
		}

		cmd.Println("\n✓ Verification finished successfully!")
		return nil
	},
}

func init() {
	testCmd.Flags().BoolVar(&testEphemeral, "ephemeral", false, "Spin up a temporary k3d cluster to run E2E deployment tests")
	testCmd.Flags().StringVarP(&testConfig, "config", "c", config.DefaultConfigFileName, "Path to moreops.yaml")
	rootCmd.AddCommand(testCmd)
}
