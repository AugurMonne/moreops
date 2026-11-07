package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"moreops/internal/core/config"
	"moreops/internal/core/diff"
	"moreops/internal/core/engine"
	"moreops/internal/core/plugin"
)

var (
	generateEnv    string
	generateConfig string
	generateModule string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate deployment and infrastructure manifests",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(generateConfig)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w (did you run 'moreops init'?)", err)
		}

		eng := engine.NewEngine(cmd.InOrStdin(), cmd.OutOrStdout())
		plan, err := eng.Plan(context.Background(), cfg, generateEnv)
		if err != nil {
			return fmt.Errorf("generate plan error: %w", err)
		}

		if generateModule != "" {
			targetMod := strings.ToLower(strings.TrimSpace(generateModule))
			filteredModulePlans := make([]*plugin.PlanResult, 0)
			filteredChanges := make([]diff.FileChange, 0)

			for _, mp := range plan.ModulePlans {
				if strings.ToLower(mp.ModuleName) == targetMod {
					filteredModulePlans = append(filteredModulePlans, mp)
					filteredChanges = append(filteredChanges, mp.FileChanges...)
				}
			}

			plan.ModulePlans = filteredModulePlans
			plan.Changes = filteredChanges
			plan.ToCreate = 0
			plan.ToUpdate = 0
			plan.ToDelete = 0
			plan.Unchanged = 0

			for _, fc := range plan.Changes {
				switch fc.Action {
				case diff.ActionCreate:
					plan.ToCreate++
				case diff.ActionUpdate:
					plan.ToUpdate++
				case diff.ActionDelete:
					plan.ToDelete++
				case diff.ActionNone:
					plan.Unchanged++
				}
			}
		}

		if len(plan.Changes) == 0 {
			cmd.Printf("No manifests to generate for module '%s' in environment '%s'.\n", generateModule, generateEnv)
			return nil
		}

		cmd.Printf("Generating %d manifest(s)...\n", len(plan.Changes))
		return eng.Apply(context.Background(), plan, true)
	},
}

func init() {
	generateCmd.Flags().StringVarP(&generateEnv, "env", "e", "dev", "Target environment name (e.g. dev, prod)")
	generateCmd.Flags().StringVarP(&generateConfig, "config", "c", config.DefaultConfigFileName, "Path to moreops.yaml")
	generateCmd.Flags().StringVarP(&generateModule, "module", "m", "", "Specific module to generate (e.g. monitoring, containers, kubernetes, security, storage, cicd)")
	rootCmd.AddCommand(generateCmd)
}
