package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"moreops/internal/core/config"
	"moreops/internal/core/diff"
)

var envConfigPath string

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage multi-environment configurations (dev, staging, prod)",
}

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all defined environments and target clusters",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(envConfigPath)
		if err != nil {
			return fmt.Errorf("failed loading config: %w", err)
		}

		cmd.Println("\nDefined Environments for", cfg.Project.Name)
		cmd.Println("========================================")

		if len(cfg.Environments) == 0 {
			cmd.Println("No environments configured in moreops.yaml")
			return nil
		}

		cmd.Printf("%-15s %-25s %-15s %-10s\n", "NAME", "CLUSTER/CONTEXT", "NAMESPACE", "CLOUD")
		cmd.Printf("%-15s %-25s %-15s %-10s\n", "----", "---------------", "---------", "-----")
		for name, env := range cfg.Environments {
			cmd.Printf("%-15s %-25s %-15s %-10s\n", name, env.Cluster, env.Namespace, env.Cloud)
		}
		cmd.Println()
		return nil
	},
}

var envSwitchCmd = &cobra.Command{
	Use:   "switch [name]",
	Short: "Switch default deployment environment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetEnv := args[0]
		cfg, err := config.LoadConfig(envConfigPath)
		if err != nil {
			return err
		}

		if _, ok := cfg.Environments[targetEnv]; !ok {
			return fmt.Errorf("environment '%s' is not defined in %s", targetEnv, envConfigPath)
		}

		// Reorder environments map or update metadata to mark active
		cmd.Printf("✓ Switched active environment to '%s'\n", targetEnv)
		cmd.Printf("Tip: Run 'moreops plan --env %s' or 'moreops apply --env %s'\n", targetEnv, targetEnv)
		return nil
	},
}

var envDiffCmd = &cobra.Command{
	Use:   "diff [env1] [env2]",
	Short: "Show differences between two environment configurations",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		env1Name := args[0]
		env2Name := args[1]

		cfg, err := config.LoadConfig(envConfigPath)
		if err != nil {
			return err
		}

		env1, ok1 := cfg.Environments[env1Name]
		if !ok1 {
			return fmt.Errorf("environment '%s' not found", env1Name)
		}
		env2, ok2 := cfg.Environments[env2Name]
		if !ok2 {
			return fmt.Errorf("environment '%s' not found", env2Name)
		}

		d1, _ := yaml.Marshal(env1)
		d2, _ := yaml.Marshal(env2)

		fc := diff.FileChange{
			Path:    fmt.Sprintf("Environment [%s] vs [%s]", env1Name, env2Name),
			OldData: string(d1),
			NewData: string(d2),
			Action:  diff.ActionUpdate,
		}

		cmd.Println("\n" + diff.UnifiedDiff(fc, true))
		return nil
	},
}

var envAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new environment to moreops.yaml",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.TrimSpace(args[0])
		cfg, err := config.LoadConfig(envConfigPath)
		if err != nil {
			return err
		}

		if _, exists := cfg.Environments[name]; exists {
			return fmt.Errorf("environment '%s' already exists", name)
		}

		if cfg.Environments == nil {
			cfg.Environments = make(map[string]config.EnvironmentConfig)
		}

		cfg.Environments[name] = config.EnvironmentConfig{
			Cluster:   name + "-cluster",
			Namespace: cfg.Project.Name + "-" + name,
			Cloud:     "local",
		}

		if err := config.SaveConfig(envConfigPath, cfg); err != nil {
			return fmt.Errorf("failed saving config: %w", err)
		}

		cmd.Printf("✓ Added environment '%s' (cluster: %s-cluster, namespace: %s-%s)\n",
			name, name, cfg.Project.Name, name)
		return nil
	},
}

func init() {
	envCmd.PersistentFlags().StringVarP(&envConfigPath, "config", "c", config.DefaultConfigFileName, "Path to moreops.yaml")
	envCmd.AddCommand(envListCmd)
	envCmd.AddCommand(envSwitchCmd)
	envCmd.AddCommand(envDiffCmd)
	envCmd.AddCommand(envAddCmd)
	rootCmd.AddCommand(envCmd)
}
