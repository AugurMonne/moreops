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
