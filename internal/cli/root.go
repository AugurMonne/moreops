package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	// Register core modules
	_ "moreops/internal/modules/cicd"
	_ "moreops/internal/modules/docker"
	_ "moreops/internal/modules/kubernetes"
	_ "moreops/internal/modules/monitoring"
	_ "moreops/internal/modules/security"
	_ "moreops/internal/modules/storage"
)

var (
	Version = "0.1.0"
	rootCmd = &cobra.Command{
		Use:   "moreops",
		Short: "Universal DevOps Orchestration Toolkit",
		Long: `MoreOps - One CLI for your entire DevOps workflow.

Manage Docker, Kubernetes, Monitoring, Vault, S3,
CI/CD and Infrastructure as Code without reinventing the wheel.`,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(initCmd)
}
