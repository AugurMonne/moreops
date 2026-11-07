package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"moreops/internal/core/config"
)

var statusConfigPath string

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the current status of the MoreOps project, modules, and manifests",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(statusConfigPath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w (run 'moreops init' first)", statusConfigPath, err)
		}

		cmd.Println("\n==========================================")
		cmd.Println("         MoreOps Project Status")
		cmd.Println("==========================================")
		cmd.Printf("Project Name:    %s\n", cfg.Project.Name)
		cmd.Printf("Project Version: %s\n", cfg.Project.Version)
		if cfg.Project.Description != "" {
			cmd.Printf("Description:     %s\n", cfg.Project.Description)
		}
		if cfg.Project.Language != "" {
			cmd.Printf("Language:        %s\n", cfg.Project.Language)
		}

		cmd.Println("\nEnvironments:")
		for envName, env := range cfg.Environments {
			cmd.Printf("  - %-10s (cluster: %s, namespace: %s)\n", envName, env.Cluster, env.Namespace)
		}

		cmd.Println("\nActive Modules:")
		if cfg.Modules.Containers != nil && cfg.Modules.Containers.Enabled {
			cmd.Printf("  [✓] Containers:    %s (compose: %t)\n", cfg.Modules.Containers.Engine, cfg.Modules.Containers.Compose)
		}
		if cfg.Modules.Kubernetes != nil && cfg.Modules.Kubernetes.Enabled {
			cmd.Printf("  [✓] Kubernetes:    tool=%s, replicas=%d, port=%d\n", cfg.Modules.Kubernetes.Tool, cfg.Modules.Kubernetes.Replicas, cfg.Modules.Kubernetes.Port)
		}
		if cfg.Modules.Monitoring != nil && cfg.Modules.Monitoring.Enabled {
			cmd.Printf("  [✓] Monitoring:    interval=%s, dashboards=%t, alert_rules=%t\n", cfg.Modules.Monitoring.ScrapeInterval, cfg.Modules.Monitoring.Dashboards, cfg.Modules.Monitoring.AlertRules)
		}
		if cfg.Modules.Security != nil && cfg.Modules.Security.Enabled {
			mode := "disabled"
			if cfg.Modules.Security.Vault != nil && cfg.Modules.Security.Vault.Enabled {
				mode = cfg.Modules.Security.Vault.Mode
			}
			cmd.Printf("  [✓] Security:      vault=%s, eso=%t\n", mode, cfg.Modules.Security.ExternalSecrets != nil && cfg.Modules.Security.ExternalSecrets.Enabled)
		}
		if cfg.Modules.Storage != nil && cfg.Modules.Storage.Enabled {
			cmd.Printf("  [✓] Storage:       provider=%s, bucket=%s\n", cfg.Modules.Storage.Provider, cfg.Modules.Storage.Bucket)
		}
		if cfg.Modules.CICD != nil && cfg.Modules.CICD.Enabled {
			gitops := "none"
			if cfg.Modules.CICD.GitOps != nil && cfg.Modules.CICD.GitOps.Enabled {
				gitops = cfg.Modules.CICD.GitOps.Tool
			}
			cmd.Printf("  [✓] CI/CD:         provider=%s, gitops=%s\n", cfg.Modules.CICD.Provider, gitops)
		}

		// Manifests check
		cmd.Println("\nManifests on Disk:")
		checkPaths := []string{"Dockerfile", "docker-compose.yml", "charts", "k8s", "Jenkinsfile", ".github", ".gitlab-ci.yml", "gitops"}
		for _, p := range checkPaths {
			if _, err := os.Stat(p); err == nil {
				cmd.Printf("  [✓] %s exists\n", p)
			}
		}

		cmd.Println()
		return nil
	},
}

func init() {
	statusCmd.Flags().StringVarP(&statusConfigPath, "config", "c", config.DefaultConfigFileName, "Path to moreops.yaml")
	rootCmd.AddCommand(statusCmd)
}
