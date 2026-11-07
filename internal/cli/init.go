package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"moreops/internal/core/config"
	"moreops/internal/wizard"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new DevOps project interactively",
	RunE: func(cmd *cobra.Command, args []string) error {
		w := wizard.NewWizard(cmd.InOrStdin(), cmd.OutOrStdout())
		cfg, err := w.Run(".")
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
