package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"moreops/internal/core/config"
	"moreops/internal/core/engine"
)

var (
	planEnv     string
	planConfig  string
	planSummary bool
)

var planCmd = &cobra.Command{
	Use:     "plan",
	Aliases: []string{"diff"},
	Short:   "Preview infrastructure and manifest changes with unified diff",
	Long: `Compute and display an execution plan showing all changes that MoreOps will make
to files, manifests, and configurations for the specified environment.

Plan performs dry-run comparison and never mutates any files on disk.`,
	Example: `  # Preview planned changes for default dev environment
  moreops plan

  # Preview high-level module and resource summary without raw diffs
  moreops plan --summary

  # Preview planned changes for production environment
  moreops plan --env prod

  # Preview using a custom config file
  moreops plan -c /path/to/moreops.yaml -e staging`,
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

		if planSummary {
			eng.PrintSummary(plan)
		} else {
			eng.PrintDiff(plan, true)
		}
		return nil
	},
}

func init() {
	planCmd.Flags().StringVarP(&planEnv, "env", "e", "dev", "Target environment name (e.g. dev, prod)")
	planCmd.Flags().StringVarP(&planConfig, "config", "c", config.DefaultConfigFileName, "Path to moreops.yaml")
	planCmd.Flags().BoolVarP(&planSummary, "summary", "s", false, "Display high-level resource summary instead of full unified diff")
	rootCmd.AddCommand(planCmd)
}
