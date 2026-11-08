package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"moreops/internal/core/config"
)

var (
	diagEnv       string
	diagNamespace string
	diagConfig    string
)

type podListOutput struct {
	Items []podItem `json:"items"`
}

type podItem struct {
	Metadata struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
	Status struct {
		Phase             string            `json:"phase"`
		ContainerStatuses []containerStatus `json:"containerStatuses"`
	} `json:"status"`
}

type containerStatus struct {
	Name         string `json:"name"`
	Ready        bool   `json:"ready"`
	RestartCount int    `json:"restartCount"`
	State        struct {
		Waiting *struct {
			Reason  string `json:"reason"`
			Message string `json:"message"`
		} `json:"waiting,omitempty"`
		Terminated *struct {
			ExitCode int    `json:"exitCode"`
			Reason   string `json:"reason"`
			Message  string `json:"message"`
		} `json:"terminated,omitempty"`
	} `json:"state"`
	LastTerminationState struct {
		Terminated *struct {
			ExitCode int    `json:"exitCode"`
			Reason   string `json:"reason"`
		} `json:"terminated,omitempty"`
	} `json:"lastState,omitempty"`
}

var diagnoseCmd = &cobra.Command{
	Use:     "diagnose",
	Aliases: []string{"troubleshoot", "debug"},
	Short:   "Diagnose failing Kubernetes workloads, pod crashes, and cluster events",
	Long: `Inspect live pods and workload health in the active Kubernetes cluster context.
Diagnose detects CrashLoopBackOff, OOMKilled containers, ImagePullBackOff errors,
and recent warning events, providing human-readable root cause explanations and fixes.`,
	Example: `  # Diagnose workloads for default dev environment
  moreops diagnose

  # Diagnose workloads in production environment
  moreops diagnose --env prod

  # Diagnose specific Kubernetes namespace
  moreops diagnose -n payments`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetNs := diagNamespace
		if targetNs == "" {
			cfg, err := config.LoadConfig(diagConfig)
			if err == nil && cfg != nil {
				if envCfg, ok := cfg.GetEnvironment(diagEnv); ok && envCfg.Namespace != "" {
					targetNs = envCfg.Namespace
				}
			}
			if targetNs == "" {
				targetNs = "default"
			}
		}

		cmd.Printf("\n=== MoreOps Kubernetes Workload Diagnostics [%s] ===\n\n", targetNs)

		// 1. Check kubectl in PATH
		if _, err := exec.LookPath("kubectl"); err != nil {
			cmd.Println("  [!] kubectl is not installed in $PATH.")
			cmd.Println("      Install kubectl or run 'moreops doctor' to inspect your toolchain.")
			return nil
		}

		// 2. Check cluster context
		ctxCmd := exec.Command("kubectl", "config", "current-context")
		var ctxOut bytes.Buffer
		ctxCmd.Stdout = &ctxOut
		if err := ctxCmd.Run(); err != nil || strings.TrimSpace(ctxOut.String()) == "" {
			cmd.Println("  [!] No active Kubernetes context found.")
			cmd.Println("      Ensure your kubeconfig (~/.kube/config) is configured and a cluster is accessible.")
			return nil
		}
		cmd.Printf("  • Cluster Context: %s\n", strings.TrimSpace(ctxOut.String()))
		cmd.Printf("  • Target Namespace: %s\n\n", targetNs)

		// 3. Inspect Pods JSON
		podCmd := exec.Command("kubectl", "get", "pods", "-n", targetNs, "-o", "json")
		var podOut bytes.Buffer
		var podErr bytes.Buffer
		podCmd.Stdout = &podOut
		podCmd.Stderr = &podErr

		if err := podCmd.Run(); err != nil {
			cmd.Printf("  [!] Failed to query pods in namespace '%s': %s\n", targetNs, strings.TrimSpace(podErr.String()))
			return nil
		}

		var pods podListOutput
		if err := json.Unmarshal(podOut.Bytes(), &pods); err != nil {
			cmd.Printf("  [!] Failed to parse pod status: %v\n", err)
			return nil
		}

		if len(pods.Items) == 0 {
			cmd.Printf("  [i] No pods found in namespace '%s'.\n", targetNs)
			cmd.Println("      Deploy manifests with 'moreops apply' or verify the namespace name.")
			return nil
		}

		issuesCount := 0
		totalPods := len(pods.Items)

		for _, pod := range pods.Items {
			podIssue := false
			var issuesList []string
			var suggestions []string

			if pod.Status.Phase != "Running" && pod.Status.Phase != "Succeeded" {
				podIssue = true
				issuesList = append(issuesList, fmt.Sprintf("Phase is '%s'", pod.Status.Phase))
			}

			for _, cs := range pod.Status.ContainerStatuses {
				if cs.RestartCount > 0 {
					podIssue = true
					issuesList = append(issuesList, fmt.Sprintf("Container '%s' has %d restart(s)", cs.Name, cs.RestartCount))
				}

				if cs.State.Waiting != nil {
					reason := cs.State.Waiting.Reason
					podIssue = true
					issuesList = append(issuesList, fmt.Sprintf("Container '%s' waiting: %s", cs.Name, reason))

					switch reason {
					case "CrashLoopBackOff":
						suggestions = append(suggestions, fmt.Sprintf("Application crashed on startup. Check logs: 'kubectl logs %s -c %s'", pod.Metadata.Name, cs.Name))
					case "ImagePullBackOff", "ErrImagePull":
						suggestions = append(suggestions, "Image could not be pulled. Verify image repository, tag, and imagePullSecrets.")
					}
				}

				if cs.State.Terminated != nil && cs.State.Terminated.ExitCode != 0 {
					podIssue = true
					reason := cs.State.Terminated.Reason
					code := cs.State.Terminated.ExitCode
					issuesList = append(issuesList, fmt.Sprintf("Container '%s' terminated (%s, exit code %d)", cs.Name, reason, code))

					if code == 137 || reason == "OOMKilled" {
						suggestions = append(suggestions, fmt.Sprintf("Container killed by OOM (Out Of Memory). Increase memory limit in 'charts/<app>/values.yaml'."))
					}
				}

				if cs.LastTerminationState.Terminated != nil && cs.LastTerminationState.Terminated.Reason == "OOMKilled" {
					podIssue = true
					issuesList = append(issuesList, fmt.Sprintf("Container '%s' was previously OOMKilled", cs.Name))
					suggestions = append(suggestions, "Container exceeded memory limit. Increase resources.limits.memory in values.yaml.")
				}
			}

			if podIssue {
				issuesCount++
				cmd.Printf("  ❌ Pod: %s\n", pod.Metadata.Name)
				for _, iss := range issuesList {
					cmd.Printf("     • Problem: %s\n", iss)
				}
				for _, sug := range suggestions {
					cmd.Printf("     💡 Recommendation: %s\n", sug)
				}
				cmd.Println()
			}
		}

		// 4. Query recent warning events
		eventCmd := exec.Command("kubectl", "get", "events", "-n", targetNs, "--field-selector", "type=Warning", "--sort-by=.lastTimestamp")
		var eventOut bytes.Buffer
		eventCmd.Stdout = &eventOut
		_ = eventCmd.Run()

		if lines := strings.Split(strings.TrimSpace(eventOut.String()), "\n"); len(lines) > 1 {
			cmd.Printf("  Recent Warning Events in namespace [%s]:\n", targetNs)
			// Print last 5 events
			start := 1
			if len(lines)-1 > 5 {
				start = len(lines) - 5
			}
			for i := start; i < len(lines); i++ {
				cmd.Printf("     ⚠️  %s\n", strings.TrimSpace(lines[i]))
			}
			cmd.Println()
		}

		if issuesCount == 0 {
			cmd.Printf("  [✓] All %d pod(s) are healthy in namespace '%s'!\n", totalPods, targetNs)
			cmd.Println("      No container crashes, OOM events, or restart loops detected.")
		} else {
			cmd.Printf("  -------------------------------------------------------\n")
			cmd.Printf("  Diagnostics found issues in %d of %d pod(s).\n\n", issuesCount, totalPods)
		}

		return nil
	},
}

func init() {
	diagnoseCmd.Flags().StringVarP(&diagEnv, "env", "e", "dev", "Target environment name (e.g. dev, prod)")
	diagnoseCmd.Flags().StringVarP(&diagNamespace, "namespace", "n", "", "Override Kubernetes namespace to diagnose")
	diagnoseCmd.Flags().StringVarP(&diagConfig, "config", "c", config.DefaultConfigFileName, "Path to moreops.yaml")
	rootCmd.AddCommand(diagnoseCmd)
}
