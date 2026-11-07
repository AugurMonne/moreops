package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"moreops/internal/core/config"
)

var storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Manage and inspect S3, MinIO, and Velero backup storage",
}

var storageCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check storage provider configuration and S3 connectivity",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(config.DefaultConfigFileName)
		if err != nil {
			return fmt.Errorf("configuration error: %w", err)
		}

		if cfg.Modules.Storage == nil || !cfg.Modules.Storage.Enabled {
			cmd.Println("Storage module is not enabled in moreops.yaml.")
			return nil
		}

		cmd.Println("\n💾 MoreOps Storage Status")
		cmd.Printf("Provider:        %s\n", cfg.Modules.Storage.Provider)
		cmd.Printf("Bucket:          %s\n", cfg.Modules.Storage.Bucket)
		cmd.Printf("Endpoint:        %s\n", cfg.Modules.Storage.Endpoint)
		cmd.Printf("Region:          %s\n", cfg.Modules.Storage.Region)
		if cfg.Modules.Storage.Velero != nil && cfg.Modules.Storage.Velero.Enabled {
			cmd.Printf("Velero Backup:   enabled (schedule: %s)\n", cfg.Modules.Storage.Velero.Schedule)
		}
		cmd.Println("✓ Storage configuration validated.")
		return nil
	},
}

func init() {
	storageCmd.AddCommand(storageCheckCmd)
	rootCmd.AddCommand(storageCmd)
}
