package wizard

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"

	"moreops/internal/core/config"
	"moreops/internal/detector"
)

// Option represents a selectable feature/component in the wizard.
type Option struct {
	ID          int
	Name        string
	Description string
}

// DefaultOptions returns the list of available modules to choose from.
func DefaultOptions() []Option {
	return []Option{
		{ID: 1, Name: "Containers & Docker", Description: "Dockerfile, Docker Compose, container builds"},
		{ID: 2, Name: "Kubernetes Manifests", Description: "Deployment, Service, ConfigMap, Ingress, HPA"},
		{ID: 3, Name: "Monitoring & Observability", Description: "Prometheus, Grafana, Loki, OpenTelemetry"},
		{ID: 4, Name: "Security & Secrets", Description: "HashiCorp Vault, External Secrets Operator, Trivy"},
		{ID: 5, Name: "Storage & Backup", Description: "S3, MinIO, Velero backup/restore"},
		{ID: 6, Name: "CI/CD & Automation", Description: "GitHub Actions, GitLab CI, Terraform/Tofu, ArgoCD"},
	}
}

// Wizard handles interactive project setup.
type Wizard struct {
	in       io.Reader
	out      io.Writer
	detector *detector.Detector
}

// NewWizard creates a Wizard with custom IO (for tests or CLI).
func NewWizard(in io.Reader, out io.Writer) *Wizard {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	return &Wizard{
		in:       in,
		out:      out,
		detector: detector.NewDetector(),
	}
}

// Ask prompts the user with a question and a default value.
func (w *Wizard) Ask(reader *bufio.Reader, question, defaultValue string) string {
	if defaultValue != "" {
		fmt.Fprintf(w.out, "%s [%s]: ", question, defaultValue)
	} else {
		fmt.Fprintf(w.out, "%s: ", question)
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}
	val := strings.TrimSpace(line)
	if val == "" {
		return defaultValue
	}
	return val
}

// ParseSelection parses user inputs such as "1 4", "1,4", "1, 4" or "14".
func ParseSelection(input string, maxOption int) []int {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	selectedMap := make(map[int]bool)

	// If input contains spaces or commas, split by them
	if strings.ContainsAny(input, " ,;") {
		parts := strings.FieldsFunc(input, func(r rune) bool {
			return r == ' ' || r == ',' || r == ';'
		})
		for _, part := range parts {
			if num, err := strconv.Atoi(part); err == nil && num >= 1 && num <= maxOption {
				selectedMap[num] = true
			}
		}
	} else {
		// Could be a single number (e.g. "1") or compact digits (e.g. "14" -> 1 and 4)
		if num, err := strconv.Atoi(input); err == nil && num >= 1 && num <= maxOption {
			selectedMap[num] = true
		} else {
			// Treat as sequence of single digits
			for _, r := range input {
				if unicode.IsDigit(r) {
					d := int(r - '0')
					if d >= 1 && d <= maxOption {
						selectedMap[d] = true
					}
				}
			}
		}
	}

	var result []int
	for i := 1; i <= maxOption; i++ {
		if selectedMap[i] {
			result = append(result, i)
		}
	}
	return result
}

// Run executes the interactive onboarding wizard.
func (w *Wizard) Run(projectDir string) (*config.Config, error) {
	reader := bufio.NewReader(w.in)

	// Step 1: Detect project
	info, _ := w.detector.Detect(projectDir)

	fmt.Fprintln(w.out, "\n========================================")
	fmt.Fprintln(w.out, "       MoreOps Project Wizard")
	fmt.Fprintln(w.out, "========================================")

	if info != nil && info.Language != "" {
		fmt.Fprintf(w.out, "🔍 Detected: %s", strings.ToUpper(info.Language))
		if info.Framework != "" {
			fmt.Fprintf(w.out, " (%s)", info.Framework)
		}
		fmt.Fprintln(w.out)
	}

	// Step 2: Basic metadata
	defaultName := "my-app"
	if info != nil && info.Name != "" {
		defaultName = info.Name
	}
	name := w.Ask(reader, "Project name", defaultName)
	version := w.Ask(reader, "Project version", "0.1.0")

	// Step 3: Environments
	envName := w.Ask(reader, "Default environment name", "dev")
	cluster := w.Ask(reader, "Kubernetes context/cluster", "current-context")
	namespace := w.Ask(reader, "Kubernetes namespace", "default")

	// Step 4: Components multi-select
	options := DefaultOptions()
	fmt.Fprintln(w.out, "\nSelect components to include:")
	for _, opt := range options {
		fmt.Fprintf(w.out, "  [%d] %-30s - %s\n", opt.ID, opt.Name, opt.Description)
	}
	fmt.Fprintln(w.out)

	selectionInput := w.Ask(reader, "Enter component numbers (e.g. 1 4, 1,4 or 14)", "1 2")
	selectedIDs := ParseSelection(selectionInput, len(options))
	if len(selectedIDs) == 0 {
		selectedIDs = []int{1, 2}
	}

	selectedSet := make(map[int]bool)
	for _, id := range selectedIDs {
		selectedSet[id] = true
	}

	// Build Config
	cfg := &config.Config{
		Version: "1",
		Project: config.ProjectMeta{
			Name:    name,
			Version: version,
		},
		Environments: map[string]config.EnvironmentConfig{
			envName: {
				Cluster:   cluster,
				Namespace: namespace,
				Cloud:     "local",
			},
		},
		Modules: config.ModulesConfig{},
	}

	if info != nil {
		cfg.Project.Language = info.Language
		cfg.Project.Framework = info.Framework
	}

	defaultPort := 8080
	if info != nil && info.DefaultPort > 0 {
		defaultPort = info.DefaultPort
	}

	// Module 1: Containers
	if selectedSet[1] {
		cfg.Modules.Containers = &config.ContainersModuleConfig{
			Enabled: true,
			Engine:  "docker",
			Image:   name + ":latest",
			Compose: true,
		}
	}

	// Module 2: Kubernetes
	if selectedSet[2] {
		k8sTool := "helm"
		toolChoice := w.Ask(reader, "Kubernetes strategy: [1] Raw Manifests, [2] Helm (Recommended), [3] Kustomize", "2")
		switch strings.TrimSpace(toolChoice) {
		case "1":
			k8sTool = "manifests"
		case "3":
			k8sTool = "kustomize"
		default:
			k8sTool = "helm"
		}

		replicasStr := w.Ask(reader, "Initial replicas count", "2")
		replicas := 2
		if r, err := strconv.Atoi(strings.TrimSpace(replicasStr)); err == nil && r > 0 {
			replicas = r
		}

		hpaChoice := w.Ask(reader, "Enable Horizontal Pod Autoscaling (HPA)? (yes/no)", "yes")
		enableHPA := strings.ToLower(strings.TrimSpace(hpaChoice)) == "yes" || strings.ToLower(strings.TrimSpace(hpaChoice)) == "y"
		minReplicas := replicas
		maxReplicas := 10
		if enableHPA {
			maxRepStr := w.Ask(reader, "HPA maximum replicas", "10")
			if m, err := strconv.Atoi(strings.TrimSpace(maxRepStr)); err == nil && m >= minReplicas {
				maxReplicas = m
			}
		}

		cfg.Modules.Kubernetes = &config.KubernetesModuleConfig{
			Enabled:  true,
			Tool:     k8sTool,
			Port:     defaultPort,
			Replicas: replicas,
			Ingress:  true,
			Autoscaling: &config.AutoscalingConfig{
				Enabled:       enableHPA,
				MinReplicas:   minReplicas,
				MaxReplicas:   maxReplicas,
				CPUPercentage: 80,
			},
		}
	}

	// Module 3: Monitoring
	if selectedSet[3] {
		monChoice := w.Ask(reader, "Observability scope: [1] Metrics & Deep Dashboards (Prometheus, Grafana, Alertmanager), [2] Full Stack (+ Loki & OpenTelemetry)", "1")
		fullStack := strings.TrimSpace(monChoice) == "2"

		cfg.Modules.Monitoring = &config.MonitoringModuleConfig{
			Enabled:        true,
			Prometheus:     true,
			Grafana:        true,
			Alertmanager:   true,
			Loki:           fullStack,
			OpenTelemetry:  fullStack,
			MetricsPath:    "/metrics",
			ScrapeInterval: "15s",
			Dashboards:     true,
			AlertRules:     true,
		}
	}

	// Module 4: Security
	if selectedSet[4] {
		vaultChoice := w.Ask(reader, "Vault setup: [1] In-cluster dev Vault (Recommended for testing), [2] External Vault", "1")
		vaultMode := "in-cluster-dev"
		vaultAddress := "http://vault.vault.svc:8200"
		if strings.TrimSpace(vaultChoice) == "2" {
			vaultMode = "external"
			vaultAddress = w.Ask(reader, "External Vault address", "https://vault.example.com")
		}

		rbacChoice := w.Ask(reader, "RBAC policy: [1] Strict least-privilege ServiceAccount (Recommended), [2] Manual, [3] Skip/Default", "1")
		rbacMode := "strict"
		switch strings.TrimSpace(rbacChoice) {
		case "2":
			rbacMode = "manual"
		case "3":
			rbacMode = "disabled"
		default:
			rbacMode = "strict"
		}

		cfg.Modules.Security = &config.SecurityModuleConfig{
			Enabled: true,
			Vault: &config.VaultConfig{
				Enabled:    true,
				Mode:       vaultMode,
				Address:    vaultAddress,
				AuthMethod: "kubernetes",
			},
			ExternalSecrets: &config.ExternalSecretsConfig{
				Enabled:         true,
				RefreshInterval: "1h",
				SecretStoreName: "vault-backend",
				TargetSecrets:   []string{name + "-secret"},
			},
			RBAC: &config.RBACConfig{
				Mode: rbacMode,
			},
			Trivy: &config.TrivyConfig{
				Enabled:  true,
				Severity: "HIGH,CRITICAL",
			},
		}
	}

	// Module 5: Storage & Backup
	if selectedSet[5] {
		provChoice := w.Ask(reader, "S3 storage provider: [1] In-cluster MinIO (Dev/Testing), [2] Yandex Cloud Object Storage, [3] VK Cloud Storage, [4] Selectel S3, [5] AWS S3", "1")
		provider := "minio-dev"
		endpoint := "http://minio:9000"
		region := "us-east-1"

		switch strings.TrimSpace(provChoice) {
		case "2":
			provider = "yandex"
			endpoint = "https://storage.yandexcloud.net"
			region = "ru-central1"
		case "3":
			provider = "vkcloud"
			endpoint = "https://hb.ru-msk.vkcs.cloud"
			region = "ru-msk"
		case "4":
			provider = "selectel"
			endpoint = "https://s3.ru-1.storage.selcloud.ru"
			region = "ru-1"
		case "5":
			provider = "aws"
			endpoint = "https://s3.amazonaws.com"
			region = "us-east-1"
		default:
			provider = "minio-dev"
			endpoint = "http://minio:9000"
			region = "us-east-1"
		}

		bucket := w.Ask(reader, "S3 backup bucket name", name+"-backups")

		// Velero question & schedule
		veleroChoice := w.Ask(reader, "Enable Velero cluster backups? (yes/no)", "yes")
		var veleroCfg *config.VeleroConfig
		if strings.ToLower(strings.TrimSpace(veleroChoice)) == "yes" || strings.ToLower(strings.TrimSpace(veleroChoice)) == "y" {
			schedule := w.Ask(reader, "Velero backup cron schedule", "0 2 * * *")
			veleroCfg = &config.VeleroConfig{
				Enabled:  true,
				Schedule: schedule,
				TTL:      "720h",
			}
		}

		// DB Backup question
		dbChoice := w.Ask(reader, "Enable automated Database backup CronJob? (yes/no)", "yes")
		var dbCfg *config.DBBackupConfig
		if strings.ToLower(strings.TrimSpace(dbChoice)) == "yes" || strings.ToLower(strings.TrimSpace(dbChoice)) == "y" {
			engineChoice := w.Ask(reader, "Database engine: [1] PostgreSQL, [2] MySQL", "1")
			engine := "postgres"
			if strings.TrimSpace(engineChoice) == "2" {
				engine = "mysql"
			}
			dbSchedule := w.Ask(reader, "Database backup cron schedule", "0 3 * * *")
			dbCfg = &config.DBBackupConfig{
				Enabled:   true,
				Engine:    engine,
				Schedule:  dbSchedule,
				Database:  name,
				SecretRef: name + "-db-credentials",
			}
		}

		cfg.Modules.Storage = &config.StorageModuleConfig{
			Enabled:        true,
			Provider:       provider,
			Endpoint:       endpoint,
			Region:         region,
			Bucket:         bucket,
			Velero:         veleroCfg,
			DatabaseBackup: dbCfg,
		}
	}

	// Module 6: CI/CD & Automation
	if selectedSet[6] {
		ciChoice := w.Ask(reader, "CI/CD provider: [1] GitHub Actions, [2] GitLab CI, [3] Jenkins", "1")
		ciProvider := "github"
		switch strings.TrimSpace(ciChoice) {
		case "2":
			ciProvider = "gitlab"
		case "3":
			ciProvider = "jenkins"
		default:
			ciProvider = "github"
		}

		runnerChoice := w.Ask(reader, "Runner environment: [1] Cloud/Hosted (e.g. ubuntu-latest), [2] Existing self-hosted runner, [3] Scaffold in-cluster runner deployment", "1")
		runnerType := "cloud"
		runnerTag := ""
		installRunner := false

		switch strings.TrimSpace(runnerChoice) {
		case "2":
			runnerType = "self-hosted"
			runnerTag = w.Ask(reader, "Runner tag/label", "self-hosted")
		case "3":
			runnerType = "self-hosted"
			installRunner = true
			runnerTag = "in-cluster-runner"
		default:
			runnerType = "cloud"
		}

		// GitOps: default is Argo CD on Enter!
		gitopsChoice := w.Ask(reader, "GitOps engine: [1] Argo CD (Recommended), [2] Flux CD, [3] None", "1")
		var gitopsCfg *config.GitOpsConfig
		chartPath := fmt.Sprintf("charts/%s", name)
		if cfg.Modules.Kubernetes != nil && cfg.Modules.Kubernetes.Tool != "helm" {
			chartPath = "k8s"
		}

		switch strings.TrimSpace(gitopsChoice) {
		case "2":
			gitopsCfg = &config.GitOpsConfig{
				Enabled:  true,
				Tool:     "flux",
				RepoURL:  "https://github.com/org/repo.git",
				Path:     chartPath,
				AutoSync: true,
			}
		case "3":
			gitopsCfg = nil
		default: // "1" or Enter
			gitopsCfg = &config.GitOpsConfig{
				Enabled:  true,
				Tool:     "argocd",
				RepoURL:  "https://github.com/org/repo.git",
				Path:     chartPath,
				AutoSync: true,
			}
		}

		// IaC
		iacChoice := w.Ask(reader, "Include OpenTofu/Terraform baseline? (yes/no)", "yes")
		var iacCfg *config.IaCConfig
		if strings.ToLower(strings.TrimSpace(iacChoice)) == "yes" || strings.ToLower(strings.TrimSpace(iacChoice)) == "y" {
			iacCfg = &config.IaCConfig{
				Enabled:  true,
				Tool:     "tofu",
				Provider: "local",
			}
		}

		// Ansible
		ansibleChoice := w.Ask(reader, "Include Ansible host bootstrap playbook? (yes/no)", "yes")
		enableAnsible := strings.ToLower(strings.TrimSpace(ansibleChoice)) == "yes" || strings.ToLower(strings.TrimSpace(ansibleChoice)) == "y"

		cfg.Modules.CICD = &config.CICDModuleConfig{
			Enabled:       true,
			Provider:      ciProvider,
			RunnerType:    runnerType,
			RunnerTag:     runnerTag,
			InstallRunner: installRunner,
			GitOps:        gitopsCfg,
			IaC:           iacCfg,
			Ansible:       enableAnsible,
		}
	}

	return cfg, nil
}
