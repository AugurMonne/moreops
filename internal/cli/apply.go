package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"moreops/internal/core/config"
	"moreops/internal/core/engine"
)

var (
	applyEnv         string
	applyConfig      string
	applyAutoApprove bool
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply infrastructure and manifest changes with preview and confirmation",
	Long: `Applies the planned changes to local files, manifests, and configurations.

MoreOps displays a unified color diff preview and asks for explicit user confirmation
before applying changes, unless the --yes (-y) auto-approve flag is specified.

Files marked with '# moreops:ignore' or '# moreops:custom' are automatically protected
and will never be overwritten.`,
	Example: `  # Preview diff and prompt for confirmation before applying
  moreops apply

  # Apply changes for production environment
  moreops apply --env prod

  # Apply changes non-interactively (ideal for CI/CD automation)
  moreops apply --yes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(applyConfig)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w (did you run 'moreops init'?)", err)
		}

		eng := engine.NewEngine(cmd.InOrStdin(), cmd.OutOrStdout())
		plan, err := eng.Plan(context.Background(), cfg, applyEnv)
		if err != nil {
			return fmt.Errorf("plan error: %w", err)
		}

		return eng.Apply(context.Background(), plan, applyAutoApprove)
	},
}

func init() {
	applyCmd.Flags().StringVarP(&applyEnv, "env", "e", "dev", "Target environment name (e.g. dev, prod)")
	applyCmd.Flags().StringVarP(&applyConfig, "config", "c", config.DefaultConfigFileName, "Path to moreops.yaml")
	applyCmd.Flags().BoolVarP(&applyAutoApprove, "yes", "y", false, "Automatically confirm and apply changes without prompt")
	rootCmd.AddCommand(applyCmd)
}
