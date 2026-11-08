package cli

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	scanSeverity string
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Security scanning for vulnerabilities and misconfigurations using Trivy",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default to scanning filesystem
		return runTrivyScan(cmd, "fs", ".")
	},
}

var scanFSCmd = &cobra.Command{
	Use:   "fs [path]",
	Short: "Scan filesystem, dependencies, and manifests for vulnerabilities and misconfigurations",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := "."
		if len(args) > 0 {
			target = args[0]
		}
		return runTrivyScan(cmd, "fs", target)
	},
}

var scanImageCmd = &cobra.Command{
	Use:   "image [image_name]",
	Short: "Scan a container image for vulnerabilities",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTrivyScan(cmd, "image", args[0])
	},
}

func runTrivyScan(cmd *cobra.Command, scanType, target string) error {
	cmd.Println("\n==========================================")
	cmd.Printf("   MoreOps Security Scanner (Trivy)\n")
	cmd.Println("==========================================")
	cmd.Printf("Target: %s (%s)\nSeverity: %s\n\n", target, scanType, scanSeverity)

	// Check if trivy is installed natively
	if trivyPath, err := exec.LookPath("trivy"); err == nil {
		cmd.Printf("✓ Using native Trivy: %s\n", trivyPath)
		trivyArgs := []string{scanType, "--severity", scanSeverity, target}
		execCmd := exec.Command("trivy", trivyArgs...)
		execCmd.Stdout = cmd.OutOrStdout()
		execCmd.Stderr = cmd.OutOrStderr()
		execCmd.Stdin = cmd.InOrStdin()
		return execCmd.Run()
	}

	// Fallback to Docker container if trivy is not installed natively
	if dockerPath, err := exec.LookPath("docker"); err == nil {
		cmd.Printf("ℹ Trivy binary not found in PATH. Running via Docker: %s\n", dockerPath)
		absTarget, err := filepath.Abs(target)
		if err != nil {
			absTarget = target
		}

		var dockerArgs []string
		if scanType == "fs" {
			dockerArgs = []string{
				"run", "--rm",
				"-v", "/var/run/docker.sock:/var/run/docker.sock",
				"-v", fmt.Sprintf("%s:/workspace", absTarget),
				"aquasec/trivy:0.51.0",
				"fs", "--severity", scanSeverity, "/workspace",
			}
		} else {
			dockerArgs = []string{
				"run", "--rm",
				"-v", "/var/run/docker.sock:/var/run/docker.sock",
				"aquasec/trivy:0.51.0",
				"image", "--severity", scanSeverity, target,
			}
		}

		execCmd := exec.Command("docker", dockerArgs...)
		execCmd.Stdout = cmd.OutOrStdout()
		execCmd.Stderr = cmd.OutOrStderr()
		execCmd.Stdin = cmd.InOrStdin()
		if err := execCmd.Run(); err != nil {
			return fmt.Errorf("docker trivy scan failed: %w", err)
		}
		return nil
	}

	return fmt.Errorf("security scanner requires either 'trivy' or 'docker' installed (run 'moreops doctor')")
}

func init() {
	scanCmd.PersistentFlags().StringVar(&scanSeverity, "severity", "HIGH,CRITICAL", "Comma-separated vulnerability severities to report (e.g. HIGH,CRITICAL)")
	scanCmd.AddCommand(scanFSCmd)
	scanCmd.AddCommand(scanImageCmd)
	rootCmd.AddCommand(scanCmd)
}
