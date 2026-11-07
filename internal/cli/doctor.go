package cli

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check your DevOps environment",
	Run: func(cmd *cobra.Command, args []string) {
		tools := []string{
			"docker",
			"k3d",
			"kubectl",
			"helm",
			"terraform",
			"tofu",
			"ansible",
			"vault",
			"trivy",
			"aws",
			"git",
		}

		fmt.Println("\nMoreOps Environment Diagnostics")
		fmt.Println("---------------------------------")

		for _, tool := range tools {
			path, err := exec.LookPath(tool)

			if err != nil {
				fmt.Printf("[ ] %-12s Not installed\n", tool)
				continue
			}

			fmt.Printf("[✓] %-12s %s\n", tool, path)
		}

		fmt.Println("\nDiagnostics completed.")
	},
}
