package cli

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
	"moreops/internal/modules/cluster"
)

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage Kubernetes clusters and ephemeral environments",
}

var clusterListCmd = &cobra.Command{
	Use:   "list",
	Short: "List local k3d clusters and current kubectl contexts",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := cluster.NewK3dManager()
		available, reason := mgr.CheckAvailable()

		cmd.Println("\nKubernetes Contexts & Clusters")
		cmd.Println("==============================")

		if available {
			clusters, err := mgr.ListClusters(context.Background())
			if err != nil {
				cmd.Printf("k3d error: %v\n", err)
			} else {
				cmd.Println("\nk3d Clusters:")
				if len(clusters) == 0 {
					cmd.Println("  (No k3d clusters found)")
				} else {
					for _, c := range clusters {
						cmd.Printf("  • %-20s (Servers: %s, Agents: %s)\n", c.Name, c.Servers, c.Agents)
					}
				}
			}
		} else {
			cmd.Printf("\nk3d not available: %s\n", reason)
		}

		// Also check kubectl contexts
		if _, err := exec.LookPath("kubectl"); err == nil {
			out, err := exec.Command("kubectl", "config", "get-contexts").Output()
			if err == nil {
				cmd.Printf("\nKubectl Contexts:\n%s\n", string(out))
			}
		}

		return nil
	},
}

var clusterCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a local k3d Kubernetes cluster",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		mgr := cluster.NewK3dManager()
		available, reason := mgr.CheckAvailable()
		if !available {
			return fmt.Errorf("cannot create cluster: %s (run 'moreops doctor')", reason)
		}

		cmd.Printf("Creating k3d cluster '%s'...\n", name)
		if err := mgr.CreateCluster(context.Background(), name, false); err != nil {
			return err
		}
		cmd.Printf("✓ Cluster '%s' created successfully! Context: k3d-%s\n", name, name)
		return nil
	},
}

var clusterDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a local k3d Kubernetes cluster",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		mgr := cluster.NewK3dManager()
		available, reason := mgr.CheckAvailable()
		if !available {
			return fmt.Errorf("cannot delete cluster: %s", reason)
		}

		cmd.Printf("Deleting k3d cluster '%s'...\n", name)
		if err := mgr.DeleteCluster(context.Background(), name); err != nil {
			return err
		}
		cmd.Printf("✓ Cluster '%s' deleted.\n", name)
		return nil
	},
}

func init() {
	clusterCmd.AddCommand(clusterListCmd)
	clusterCmd.AddCommand(clusterCreateCmd)
	clusterCmd.AddCommand(clusterDeleteCmd)
	rootCmd.AddCommand(clusterCmd)
}
