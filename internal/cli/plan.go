package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"moreops/internal/core/config"
	"moreops/internal/core/engine"
)

var (
	planEnv    string
	planConfig string
)

var planCmd = &cobra.Command{
	Use:     "plan",
	Aliases: []string{"diff"},
	Short:   "Preview infrastructure and manifest changes with unified diff",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(planConfig)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w (did you run 'moreops init'?)", err)
		}

		eng := engine.NewEngine(cmd.InOrStdin(), cmd.OutOrStdout())
		plan, err := eng.Plan(context.Background(), cfg, planEnv)
		if err != nil {
			return fmt.Errorf("plan error: %w", err)
		}

		eng.PrintDiff(plan, true)
		return nil
	},
}

func init() {
	planCmd.Flags().StringVarP(&planEnv, "env", "e", "dev", "Target environment name (e.g. dev, prod)")
	planCmd.Flags().StringVarP(&planConfig, "config", "c", config.DefaultConfigFileName, "Path to moreops.yaml")
	rootCmd.AddCommand(planCmd)
}
