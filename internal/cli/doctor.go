package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Comprehensive diagnostics of DevOps environment and security posture",
	Long: `Diagnose your local DevOps environment, required tools, container runtimes,
cluster reachability, and project security hygiene.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("\n=======================================================")
		cmd.Println("         MoreOps DevOps Environment Diagnostics        ")
		cmd.Println("=======================================================")

		passed := 0
		warnings := 0
		errors := 0

		// 1. Toolchain & Runtimes
		cmd.Println("\n[1] Toolchain & CLI Utilities:")

		// Docker
		if path, err := exec.LookPath("docker"); err == nil {
			var daemonStatus string
			c := exec.Command("docker", "info", "--format", "{{.ServerVersion}}")
			var out bytes.Buffer
			c.Stdout = &out
			if err := c.Run(); err == nil && len(strings.TrimSpace(out.String())) > 0 {
				daemonStatus = fmt.Sprintf("daemon active (v%s)", strings.TrimSpace(out.String()))
				cmd.Printf("  [✓] %-14s %s (%s)\n", "docker", path, daemonStatus)
				passed++
			} else {
				cmd.Printf("  [!] %-14s %s (daemon not running)\n", "docker", path)
				warnings++
			}
		} else {
			cmd.Printf("  [✗] %-14s Not found in $PATH\n", "docker")
			errors++
		}

		// kubectl
		if path, err := exec.LookPath("kubectl"); err == nil {
			clusterCtx := "no context"
			c := exec.Command("kubectl", "config", "current-context")
			var out bytes.Buffer
			c.Stdout = &out
			if err := c.Run(); err == nil && len(strings.TrimSpace(out.String())) > 0 {
				clusterCtx = strings.TrimSpace(out.String())
				cmd.Printf("  [✓] %-14s %s (active context: %s)\n", "kubectl", path, clusterCtx)
				passed++
			} else {
				cmd.Printf("  [!] %-14s %s (no active cluster context)\n", "kubectl", path)
				warnings++
			}
		} else {
			cmd.Printf("  [!] %-14s Not found in $PATH (needed for live cluster ops)\n", "kubectl")
			warnings++
		}

		// Helm
		if path, err := exec.LookPath("helm"); err == nil {
			cmd.Printf("  [✓] %-14s %s\n", "helm", path)
			passed++
		} else {
			cmd.Printf("  [✓] %-14s Native fallback enabled (pure Go packaging available)\n", "helm")
			passed++
		}

		// Git
		if path, err := exec.LookPath("git"); err == nil {
			gitRepo := "outside git repo"
			c := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
			var out bytes.Buffer
			c.Stdout = &out
			if err := c.Run(); err == nil && len(strings.TrimSpace(out.String())) > 0 {
				gitRepo = fmt.Sprintf("branch: %s", strings.TrimSpace(out.String()))
			}
			cmd.Printf("  [✓] %-14s %s (%s)\n", "git", path, gitRepo)
			passed++
		} else {
			cmd.Printf("  [✗] %-14s Not found in $PATH\n", "git")
			errors++
		}

		// Optional tools
		optionalTools := []string{"tofu", "terraform", "vault", "trivy", "sops", "minio"}
		for _, tool := range optionalTools {
			if path, err := exec.LookPath(tool); err == nil {
				cmd.Printf("  [✓] %-14s %s\n", tool, path)
				passed++
			} else {
				cmd.Printf("  [ ] %-14s (optional) not installed\n", tool)
			}
		}

		// 2. Project Configuration & Security Hygiene
		cmd.Println("\n[2] Project & Security Hygiene:")

		// Config file check
		cfgFile := "moreops.yaml"
		if data, err := os.ReadFile(cfgFile); err == nil {
			var parsed map[string]any
			if err := yaml.Unmarshal(data, &parsed); err == nil {
				cmd.Printf("  [✓] Configuration:  '%s' exists and is valid YAML\n", cfgFile)
				passed++
			} else {
				cmd.Printf("  [✗] Configuration:  '%s' has YAML syntax errors: %v\n", cfgFile, err)
				errors++
			}
		} else {
			cmd.Printf("  [!] Configuration:  '%s' not found (run 'moreops init' to create)\n", cfgFile)
			warnings++
		}

		// Plaintext secret leak detection in config files
		secretLeakFound := false
		secretPatterns := []*regexp.Regexp{
			regexp.MustCompile(`(?i)(password|secret_key|api_key|private_key)\s*:\s*["']?([A-Za-z0-9_\-\.\$\!\#]{8,})["']?`),
			regexp.MustCompile(`(?i)ghp_[A-Za-z0-9]{36}`),
			regexp.MustCompile(`(?i)AKIA[0-9A-Z]{16}`),
		}

		filesToScan := []string{"moreops.yaml", ".env"}
		for _, f := range filesToScan {
			if content, err := os.ReadFile(f); err == nil {
				for _, pat := range secretPatterns {
					if pat.Match(content) {
						// Check if it's not a template expression
						if !strings.Contains(string(content), "{{") {
							cmd.Printf("  [!] Security Leak:   Potential plaintext secret found in '%s'\n", f)
							secretLeakFound = true
							warnings++
							break
						}
					}
				}
			}
		}
		if !secretLeakFound {
			cmd.Println("  [✓] Secret Hygiene: No plaintext tokens or private keys found in configs")
			passed++
		}

		// File permissions security check
		sensitiveFiles := []string{".moreops/secrets.key", ".env"}
		for _, sf := range sensitiveFiles {
			if info, err := os.Stat(sf); err == nil {
				mode := info.Mode().Perm()
				if mode&0077 != 0 {
					if doctorFix {
						if err := os.Chmod(sf, 0600); err == nil {
							cmd.Printf("  [✓] File Security:  Fixed permissions on '%s' -> 0600\n", sf)
							passed++
						} else {
							cmd.Printf("  [!] File Security:  Failed to fix permissions on '%s': %v\n", sf, err)
							warnings++
						}
					} else {
						cmd.Printf("  [!] File Security:  '%s' has loose permissions (%04o). Recommended: 0600 (Run 'moreops doctor --fix')\n", sf, mode)
						warnings++
					}
				} else {
					cmd.Printf("  [✓] File Security:  '%s' permissions restricted (%04o)\n", sf, mode)
					passed++
				}
			}
		}

		// Gitignore check
		giData, err := os.ReadFile(".gitignore")
		giStr := string(giData)
		if err == nil && (strings.Contains(giStr, ".env") || strings.Contains(giStr, ".moreops")) {
			cmd.Println("  [✓] Git Protection: Sensitive patterns (.env, .moreops) ignored in .gitignore")
			passed++
		} else {
			if doctorFix {
				f, err := os.OpenFile(".gitignore", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
				if err == nil {
					f.WriteString("\n# MoreOps secrets and environment files\n.moreops/\n.env\n.env.*\n")
					f.Close()
					cmd.Println("  [✓] Git Protection: Fixed .gitignore (added .moreops/, .env)")
					passed++
				} else {
					cmd.Printf("  [!] Git Protection: Failed to update .gitignore: %v\n", err)
					warnings++
				}
			} else {
				cmd.Println("  [!] Git Protection: .gitignore missing entry for '.moreops' or '.env' (Run 'moreops doctor --fix')")
				warnings++
			}
		}

		// Summary
		cmd.Println("\n-------------------------------------------------------")
		cmd.Printf("Diagnostics completed: %d passed, %d warning(s), %d error(s)\n", passed, warnings, errors)
		if errors == 0 && warnings == 0 {
			cmd.Println("✓ Your environment is fully ready for production DevOps workflows!")
		} else if errors > 0 {
			cmd.Println("✗ Some required tools or configs need attention before proceeding.")
		}
		cmd.Println()
	},
}

var doctorFix bool

func init() {
	doctorCmd.Flags().BoolVarP(&doctorFix, "fix", "f", false, "Automatically fix file permissions and git security hygiene issues")
}
