package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"moreops/internal/core/config"
)

var (
	statusConfigPath string
	statusLive       bool
	statusEnv        string
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the current status of the MoreOps project, modules, and manifests",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(statusConfigPath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w (run 'moreops init' first)", statusConfigPath, err)
		}

		cmd.Println("\n==========================================")
		cmd.Println("         MoreOps Project Status")
		cmd.Println("==========================================")
		cmd.Printf("Project Name:    %s\n", cfg.Project.Name)
		cmd.Printf("Project Version: %s\n", cfg.Project.Version)
		if cfg.Project.Description != "" {
			cmd.Printf("Description:     %s\n", cfg.Project.Description)
		}
		if cfg.Project.Language != "" {
			cmd.Printf("Language:        %s\n", cfg.Project.Language)
		}

		cmd.Println("\nEnvironments:")
		for envName, env := range cfg.Environments {
			cmd.Printf("  - %-10s (cluster: %s, namespace: %s)\n", envName, env.Cluster, env.Namespace)
		}

		cmd.Println("\nActive Modules:")
		if cfg.Modules.Containers != nil && cfg.Modules.Containers.Enabled {
			cmd.Printf("  [✓] Containers:    %s (compose: %t)\n", cfg.Modules.Containers.Engine, cfg.Modules.Containers.Compose)
		}
		if cfg.Modules.Kubernetes != nil && cfg.Modules.Kubernetes.Enabled {
			hpaInfo := ""
			if cfg.Modules.Kubernetes.Autoscaling != nil && cfg.Modules.Kubernetes.Autoscaling.Enabled {
				hpaInfo = fmt.Sprintf(", hpa: %d-%d", cfg.Modules.Kubernetes.Replicas, cfg.Modules.Kubernetes.Autoscaling.MaxReplicas)
			}
			cmd.Printf("  [✓] Kubernetes:    tool=%s, replicas=%d%s, port=%d\n", cfg.Modules.Kubernetes.Tool, cfg.Modules.Kubernetes.Replicas, hpaInfo, cfg.Modules.Kubernetes.Port)
		}
		if cfg.Modules.Monitoring != nil && cfg.Modules.Monitoring.Enabled {
			cmd.Printf("  [✓] Monitoring:    interval=%s, dashboards=%t, alert_rules=%t\n", cfg.Modules.Monitoring.ScrapeInterval, cfg.Modules.Monitoring.Dashboards, cfg.Modules.Monitoring.AlertRules)
		}
		if cfg.Modules.Security != nil && cfg.Modules.Security.Enabled {
			mode := "disabled"
			if cfg.Modules.Security.Vault != nil && cfg.Modules.Security.Vault.Enabled {
				mode = cfg.Modules.Security.Vault.Mode
			}
			cmd.Printf("  [✓] Security:      vault=%s, eso=%t\n", mode, cfg.Modules.Security.ExternalSecrets != nil && cfg.Modules.Security.ExternalSecrets.Enabled)
		}
		if cfg.Modules.Storage != nil && cfg.Modules.Storage.Enabled {
			cmd.Printf("  [✓] Storage:       provider=%s, bucket=%s\n", cfg.Modules.Storage.Provider, cfg.Modules.Storage.Bucket)
		}
		if cfg.Modules.CICD != nil && cfg.Modules.CICD.Enabled {
			gitops := "none"
			if cfg.Modules.CICD.GitOps != nil && cfg.Modules.CICD.GitOps.Enabled {
				gitops = cfg.Modules.CICD.GitOps.Tool
			}
			cmd.Printf("  [✓] CI/CD:         provider=%s, gitops=%s\n", cfg.Modules.CICD.Provider, gitops)
		}

		// Manifests check
		cmd.Println("\nManifests on Disk:")
		checkPaths := []string{"Dockerfile", "docker-compose.yml", "charts", "k8s", "Jenkinsfile", ".github", ".gitlab-ci.yml", "gitops"}
		for _, p := range checkPaths {
			if _, err := os.Stat(p); err == nil {
				cmd.Printf("  [✓] %s exists\n", p)
			}
		}

		// Live cluster inspection
		if statusLive {
			targetEnv := statusEnv
			if targetEnv == "" {
				targetEnv = "dev"
			}
			envCfg, exists := cfg.GetEnvironment(targetEnv)
			ns := "default"
			if exists && envCfg.Namespace != "" {
				ns = envCfg.Namespace
			}

			cmd.Printf("\n--- Live Kubernetes Cluster State [%s / namespace: %s] ---\n", targetEnv, ns)
			inspectLiveCluster(cmd, cfg.Project.Name, ns)
		} else {
			cmd.Println("\nTip: Run 'moreops status --live' to inspect live pods and workloads in your active Kubernetes cluster.")
		}

		cmd.Println()
		return nil
	},
}

func inspectLiveCluster(cmd *cobra.Command, appName, namespace string) {
	if _, err := exec.LookPath("kubectl"); err != nil {
		cmd.Println("  [!] kubectl is not installed in PATH. Cannot inspect live cluster.")
		return
	}

	// 1. Check cluster connection
	ctxCmd := exec.Command("kubectl", "config", "current-context")
	var ctxOut bytes.Buffer
	ctxCmd.Stdout = &ctxOut
	if err := ctxCmd.Run(); err != nil {
		cmd.Println("  [!] Unable to connect to Kubernetes cluster: no active context.")
		return
	}
	currentCtx := strings.TrimSpace(ctxOut.String())
	cmd.Printf("  Context:   %s\n", currentCtx)

	// 2. Check namespace
	nsCmd := exec.Command("kubectl", "get", "namespace", namespace, "-o", "jsonpath={.status.phase}")
	var nsOut bytes.Buffer
	nsCmd.Stdout = &nsOut
	if err := nsCmd.Run(); err != nil {
		cmd.Printf("  Namespace: '%s' (not found or inaccessible)\n", namespace)
	} else {
		cmd.Printf("  Namespace: '%s' (status: %s)\n", namespace, strings.TrimSpace(nsOut.String()))
	}

	// 3. Inspect Pods
	podCmd := exec.Command("kubectl", "get", "pods", "-n", namespace,
		"-l", fmt.Sprintf("app.kubernetes.io/name=%s", appName),
		"-o", "custom-columns=NAME:.metadata.name,READY:.status.containerStatuses[0].ready,STATUS:.status.phase,RESTARTS:.status.containerStatuses[0].restartCount",
		"--no-headers")
	var podOut bytes.Buffer
	podCmd.Stdout = &podOut
	_ = podCmd.Run()

	podLines := strings.TrimSpace(podOut.String())
	if podLines == "" {
		cmd.Printf("  Pods:      No running pods found with label app.kubernetes.io/name=%s\n", appName)
	} else {
		lines := strings.Split(podLines, "\n")
		cmd.Printf("  Pods:      %d found:\n", len(lines))
		for _, line := range lines {
			cmd.Printf("    • %s\n", line)
		}
	}

	// 4. Check HPA
	hpaCmd := exec.Command("kubectl", "get", "hpa", "-n", namespace, appName, "--no-headers")
	var hpaOut bytes.Buffer
	hpaCmd.Stdout = &hpaOut
	if err := hpaCmd.Run(); err == nil && len(hpaOut.String()) > 0 {
		cmd.Printf("  HPA:       %s\n", strings.TrimSpace(hpaOut.String()))
	}
}

func init() {
	statusCmd.Flags().StringVarP(&statusConfigPath, "config", "c", config.DefaultConfigFileName, "Path to moreops.yaml")
	statusCmd.Flags().BoolVarP(&statusLive, "live", "l", false, "Query live Kubernetes cluster for pod and workload status")
	statusCmd.Flags().StringVarP(&statusEnv, "env", "e", "dev", "Target environment for live cluster query")
	rootCmd.AddCommand(statusCmd)
}
