package cli

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var chartCmd = &cobra.Command{
	Use:   "chart",
	Short: "Package and push Helm charts to OCI registries",
}

var chartPackageCmd = &cobra.Command{
	Use:   "package [path]",
	Short: "Package a Helm chart directory into a versioned chart archive",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		chartPath := "charts"
		if len(args) > 0 {
			chartPath = args[0]
		}

		cmd.Printf("📦 Packaging Helm chart from '%s'...\n", chartPath)
		execCmd := exec.Command("helm", "package", chartPath)
		out, err := execCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("helm package failed: %w (output: %s)", err, string(out))
		}

		cmd.Print(string(out))
		cmd.Println("✓ Chart packaged successfully.")
		return nil
	},
}

var chartPushCmd = &cobra.Command{
	Use:   "push <archive.tgz> <oci://registry/path>",
	Short: "Push a packaged Helm chart archive to an OCI registry",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		archive := args[0]
		registry := args[1]

		cmd.Printf("🚀 Pushing chart '%s' to OCI registry '%s'...\n", archive, registry)
		execCmd := exec.Command("helm", "push", archive, registry)
		out, err := execCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("helm push failed: %w (output: %s)", err, string(out))
		}

		cmd.Print(string(out))
		cmd.Println("✓ Chart pushed successfully to OCI registry.")
		return nil
	},
}

func init() {
	chartCmd.AddCommand(chartPackageCmd)
	chartCmd.AddCommand(chartPushCmd)
	rootCmd.AddCommand(chartCmd)
}
