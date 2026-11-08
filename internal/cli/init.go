package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"moreops/internal/core/config"
	"moreops/internal/wizard"
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new DevOps project interactively",
	Long: `Initialize a standardized, production-ready DevOps configuration for your project.

MoreOps auto-detects programming languages and frameworks (Go, Node.js, Python, Java, etc.)
and guides you step-by-step through configuring:
  • Multi-stage Dockerfile and docker-compose.yml
  • Production Helm chart or Kustomize manifests with HPA
  • Monitoring (Prometheus ServiceMonitor, Grafana dashboards)
  • Security & Secrets (Vault, ESO, strict RBAC)
  • Database topologies (Shared/Managed Cloud DB vs Dedicated In-Cluster)
  • Pre-upgrade Database Migration Jobs (Helm hooks)
  • Message Brokers (RabbitMQ AMQP)
  • CI/CD and GitOps pipelines (GitLab CI, GitHub Actions, Jenkins, Argo CD, OpenTofu)`,
	Example: `  # Initialize current directory interactively
  moreops init

  # Initialize a specific project path
  moreops init ./my-service`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir := "."
		if len(args) > 0 {
			targetDir = args[0]
		}
		w := wizard.NewWizard(cmd.InOrStdin(), cmd.OutOrStdout())
		cfg, err := w.Run(targetDir)
		if err != nil {
			return err
		}

		configPath := config.DefaultConfigFileName
		if err := config.SaveConfig(configPath, cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		cmd.Println("\n✓ Configuration saved to", configPath)
		cmd.Println("\nNext steps:")
		cmd.Println("  moreops plan         # Preview planned changes with unified diff")
		cmd.Println("  moreops apply        # Review and apply changes safely")
		cmd.Println("  moreops validate     # Validate configuration and manifests")
		return nil
	},
}
