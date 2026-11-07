package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var (
	iacAutoApprove bool
)

var iacCmd = &cobra.Command{
	Use:   "iac",
	Short: "Manage Infrastructure as Code (OpenTofu / Terraform)",
}

var iacPlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "Preview infrastructure changes using OpenTofu or Terraform",
	RunE: func(cmd *cobra.Command, args []string) error {
		tool, err := findIaCTool()
		if err != nil {
			return err
		}

		if _, err := os.Stat("terraform"); os.IsNotExist(err) {
			return fmt.Errorf("terraform directory not found (run 'moreops init' with CI/CD & IaC enabled)")
		}

		cmd.Printf("Running '%s plan' in terraform/...\n\n", tool)
		runCmd := exec.Command(tool, "plan")
		runCmd.Dir = "terraform"
		runCmd.Stdout = cmd.OutOrStdout()
		runCmd.Stderr = cmd.OutOrStderr()
		runCmd.Stdin = cmd.InOrStdin()
		return runCmd.Run()
	},
}

var iacApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply infrastructure changes with OpenTofu or Terraform",
	RunE: func(cmd *cobra.Command, args []string) error {
		tool, err := findIaCTool()
		if err != nil {
			return err
		}

		if _, err := os.Stat("terraform"); os.IsNotExist(err) {
			return fmt.Errorf("terraform directory not found (run 'moreops init' with CI/CD & IaC enabled)")
		}

		if !iacAutoApprove {
			cmd.Printf("Do you want to apply OpenTofu/Terraform infrastructure changes? (yes/no): ")
			reader := bufio.NewReader(cmd.InOrStdin())
			line, _ := reader.ReadString('\n')
			ans := strings.ToLower(strings.TrimSpace(line))
			if ans != "yes" && ans != "y" {
				cmd.Println("IaC apply cancelled by user.")
				return nil
			}
		}

		cmd.Printf("Executing '%s apply -auto-approve' in terraform/...\n\n", tool)
		runCmd := exec.Command(tool, "apply", "-auto-approve")
		runCmd.Dir = "terraform"
		runCmd.Stdout = cmd.OutOrStdout()
		runCmd.Stderr = cmd.OutOrStderr()
		runCmd.Stdin = cmd.InOrStdin()
		return runCmd.Run()
	},
}

func findIaCTool() (string, error) {
	if path, err := exec.LookPath("tofu"); err == nil {
		return path, nil
	}
	if path, err := exec.LookPath("terraform"); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("neither 'tofu' nor 'terraform' found in PATH (run 'moreops doctor')")
}

func init() {
	iacApplyCmd.Flags().BoolVarP(&iacAutoApprove, "yes", "y", false, "Skip interactive confirmation")
	iacCmd.AddCommand(iacPlanCmd)
	iacCmd.AddCommand(iacApplyCmd)
	rootCmd.AddCommand(iacCmd)
}
