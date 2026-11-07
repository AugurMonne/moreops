package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"moreops/internal/core/config"
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Manage and inspect Prometheus and Grafana monitoring",
}

var monitorStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check monitoring configuration, scraping interval, and targets",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(config.DefaultConfigFileName)
		if err != nil {
			return fmt.Errorf("configuration error: %w", err)
		}

		if cfg.Modules.Monitoring == nil || !cfg.Modules.Monitoring.Enabled {
			cmd.Println("Monitoring module is not enabled in moreops.yaml.")
			return nil
		}

		cmd.Println("\n📊 MoreOps Monitoring Status")
		cmd.Printf("Metrics Path:    %s\n", cfg.Modules.Monitoring.MetricsPath)
		cmd.Printf("Scrape Interval: %s\n", cfg.Modules.Monitoring.ScrapeInterval)
		cmd.Printf("Prometheus:      %t\n", cfg.Modules.Monitoring.Prometheus)
		cmd.Printf("Grafana:         %t\n", cfg.Modules.Monitoring.Grafana)
		cmd.Printf("Alert Rules:     %t\n", cfg.Modules.Monitoring.AlertRules)
		cmd.Printf("Deep Dashboards: %t\n", cfg.Modules.Monitoring.Dashboards)
		cmd.Println("✓ Monitoring endpoints configured.")
		return nil
	},
}

func init() {
	monitorCmd.AddCommand(monitorStatusCmd)
	rootCmd.AddCommand(monitorCmd)
}
