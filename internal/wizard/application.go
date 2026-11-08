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

const (
	ansiReset   = "\033[0m"
	ansiBold    = "\033[1m"
	ansiDim     = "\033[2m"
	ansiCyan    = "\033[36m"
	ansiGreen   = "\033[32m"
	ansiYellow  = "\033[33m"
	ansiBlue    = "\033[34m"
	ansiMagenta = "\033[35m"
	ansiWhite   = "\033[37m"
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
		{ID: 1, Name: "📦  Containers & Docker", Description: "Multi-stage Dockerfile, .dockerignore, docker-compose.yml"},
		{ID: 2, Name: "☸️   Kubernetes Manifests", Description: "Helm chart (Recommended), Kustomize overlays, Raw manifests"},
		{ID: 3, Name: "📊  Monitoring & Observability", Description: "Prometheus Operator, ServiceMonitor, Grafana dashboards"},
		{ID: 4, Name: "🔐  Security & Secrets", Description: "HashiCorp Vault (in-cluster dev / external), ESO, strict RBAC"},
		{ID: 5, Name: "💾  Storage & Backup", Description: "MinIO dev pod, Cloud S3 (Yandex, VK, Selectel, AWS), Velero"},
		{ID: 6, Name: "🚀  CI/CD & GitOps Automation", Description: "Jenkins, GitHub Actions, GitLab CI, Argo CD (default), OpenTofu"},
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

func (w *Wizard) printStep(step, total int, title string) {
	dashes := 45 - len(title)
	if dashes < 5 {
		dashes = 5
	}
	fmt.Fprintf(w.out, "\n%s─── Step %d/%d: %s %s%s\n\n",
		ansiCyan+ansiBold, step, total, title,
		strings.Repeat("─", dashes), ansiReset)
}

// Ask prompts the user with a question and a default value.
func (w *Wizard) Ask(reader *bufio.Reader, question, defaultValue string) string {
	if defaultValue != "" {
		fmt.Fprintf(w.out, "%s?%s %s%s%s %s[%s]%s: ",
			ansiCyan, ansiReset,
			ansiBold, question, ansiReset,
			ansiDim, defaultValue, ansiReset)
	} else {
		fmt.Fprintf(w.out, "%s?%s %s%s%s: ",
			ansiCyan, ansiReset,
			ansiBold, question, ansiReset)
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

	fmt.Fprintln(w.out, "\n"+ansiCyan+ansiBold+"╭──────────────────────────────────────────────────────────╮"+ansiReset)
	fmt.Fprintln(w.out, ansiCyan+ansiBold+"│                 ✨  MoreOps Project Setup  ✨            │"+ansiReset)
	fmt.Fprintln(w.out, ansiCyan+ansiBold+"│         Orchestrate Your Complete DevOps Workflow        │"+ansiReset)
	fmt.Fprintln(w.out, ansiCyan+ansiBold+"╰──────────────────────────────────────────────────────────╯"+ansiReset)

	if info != nil && info.Language != "" {
		fmt.Fprintf(w.out, "\n  %s🔍 Auto-detected:%s %s%s%s", ansiGreen+ansiBold, ansiReset, ansiBold+ansiWhite, strings.ToUpper(info.Language), ansiReset)
		if info.Framework != "" {
			fmt.Fprintf(w.out, " %s(%s)%s", ansiDim, info.Framework, ansiReset)
		}
		fmt.Fprintln(w.out)
	}

	// Step 1: Metadata
	w.printStep(1, 4, "Project Metadata")
	defaultName := "my-app"
	if info != nil && info.Name != "" {
		defaultName = info.Name
	}
	name := w.Ask(reader, "Project name", defaultName)
	version := w.Ask(reader, "Project version", "0.1.0")

	// Step 2: Environments
	w.printStep(2, 4, "Target Environment")
	envName := w.Ask(reader, "Default environment name", "dev")
	cluster := w.Ask(reader, "Kubernetes context/cluster", "current-context")
	namespace := w.Ask(reader, "Kubernetes namespace", "default")

	// Step 3: Components multi-select
	w.printStep(3, 4, "Select Architecture Modules")
	options := DefaultOptions()
	var choiceOptions []Choice
	for _, opt := range options {
		choiceOptions = append(choiceOptions, Choice{
			ID:          opt.ID,
			Key:         strconv.Itoa(opt.ID),
			Label:       opt.Name,
			Description: opt.Description,
		})
	}

	selectedChoices, _ := SelectMulti(reader, w.in, w.out, "Select components to orchestrate", choiceOptions, []int{0, 1})
	selectedSet := make(map[int]bool)
	for _, sc := range selectedChoices {
		selectedSet[sc.ID] = true
	}
	if len(selectedSet) == 0 {
		selectedSet[1] = true
		selectedSet[2] = true
	}

	w.printStep(4, 4, "Module Configuration & Tuning")

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
		k8sChoices := []Choice{
			{ID: 1, Key: "helm", Label: "Helm (Recommended)", Description: "Full Helm chart with values overlays"},
			{ID: 2, Key: "manifests", Label: "Raw Manifests", Description: "Standard standalone YAMLs in k8s/"},
			{ID: 3, Key: "kustomize", Label: "Kustomize", Description: "Overlay-based manifests with kustomization.yaml"},
		}
		ch, _ := SelectOne(reader, w.in, w.out, "Kubernetes strategy", k8sChoices, 0)
		k8sTool := ch.Key

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

		ingressChoices := []Choice{
			{ID: 1, Key: "nginx", Label: "NGINX (Recommended)", Description: "Production standard high-performance controller"},
			{ID: 2, Key: "traefik", Label: "Traefik", Description: "Cloud-native edge router with dynamic discovery"},
			{ID: 3, Key: "caddy", Label: "Caddy", Description: "Automatic HTTPS and modern lightweight proxy"},
			{ID: 4, Key: "haproxy", Label: "HAProxy", Description: "Ultra high-load L4/L7 load balancer"},
			{ID: 5, Key: "none", Label: "None / Custom", Description: "Skip automated Ingress generation"},
		}
		ingChoice, _ := SelectOne(reader, w.in, w.out, "Ingress controller", ingressChoices, 0)
		ingressClass := ingChoice.Key
		enableIngress := true
		if ingChoice.Key == "none" {
			customChoice := w.Ask(reader, "Enable custom Ingress? (yes/no)", "no")
			enableIngress = strings.ToLower(strings.TrimSpace(customChoice)) == "yes" || strings.ToLower(strings.TrimSpace(customChoice)) == "y"
			if enableIngress {
				ingressClass = w.Ask(reader, "Custom Ingress class name", "custom")
			}
		}

		cfg.Modules.Kubernetes = &config.KubernetesModuleConfig{
			Enabled:      true,
			Tool:         k8sTool,
			Port:         defaultPort,
			Replicas:     replicas,
			Ingress:      enableIngress,
			IngressClass: ingressClass,
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
		monChoices := []Choice{
			{ID: 1, Key: "metrics", Label: "Metrics & Dashboards (Recommended)", Description: "Prometheus Operator, Grafana, Alertmanager"},
			{ID: 2, Key: "full", Label: "Full Stack Observability", Description: "Metrics + Loki Logs & OpenTelemetry Traces"},
		}
		mChoice, _ := SelectOne(reader, w.in, w.out, "Observability scope", monChoices, 0)
		fullStack := mChoice.Key == "full"

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
		vaultChoices := []Choice{
			{ID: 1, Key: "in-cluster-dev", Label: "In-cluster dev Vault (Recommended for testing)", Description: "Zero-config dev instance with auto-unseal"},
			{ID: 2, Key: "external", Label: "External Production Vault", Description: "Connect to enterprise HashiCorp Vault"},
		}
		vChoice, _ := SelectOne(reader, w.in, w.out, "Vault deployment mode", vaultChoices, 0)
		vaultMode := vChoice.Key
		vaultAddress := "http://vault.vault.svc:8200"
		if vaultMode == "external" {
			vaultAddress = w.Ask(reader, "External Vault address", "https://vault.example.com")
		}

		rbacChoices := []Choice{
			{ID: 1, Key: "strict", Label: "Strict least-privilege (Recommended)", Description: "Dedicated ServiceAccount with minimal RBAC permissions"},
			{ID: 2, Key: "manual", Label: "Manual", Description: "Developer-managed Role & RoleBinding"},
			{ID: 3, Key: "disabled", Label: "Skip / Default", Description: "Default namespace permissions"},
		}
		rbacChoice, _ := SelectOne(reader, w.in, w.out, "RBAC policy", rbacChoices, 0)
		rbacMode := rbacChoice.Key

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
		provChoices := []Choice{
			{ID: 1, Key: "minio-dev", Label: "In-cluster MinIO (Recommended for Dev/Testing)", Description: "Local S3 compatible storage"},
			{ID: 2, Key: "yandex", Label: "Yandex Cloud Object Storage", Description: "ru-central1 managed S3"},
			{ID: 3, Key: "vkcloud", Label: "VK Cloud Storage", Description: "ru-msk managed S3"},
			{ID: 4, Key: "selectel", Label: "Selectel S3", Description: "ru-1 managed S3"},
			{ID: 5, Key: "aws", Label: "AWS S3", Description: "us-east-1 Amazon S3"},
		}
		pChoice, _ := SelectOne(reader, w.in, w.out, "S3 storage provider", provChoices, 0)
		provider := pChoice.Key
		endpoint := "http://minio:9000"
		region := "us-east-1"

		switch provider {
		case "yandex":
			endpoint = "https://storage.yandexcloud.net"
			region = "ru-central1"
		case "vkcloud":
			endpoint = "https://hb.ru-msk.vkcs.cloud"
			region = "ru-msk"
		case "selectel":
			endpoint = "https://s3.ru-1.storage.selcloud.ru"
			region = "ru-1"
		case "aws":
			endpoint = "https://s3.amazonaws.com"
			region = "us-east-1"
		default:
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
		ciChoices := []Choice{
			{ID: 1, Key: "github", Label: "GitHub Actions (Recommended)", Description: "Declarative workflows in .github/workflows/"},
			{ID: 2, Key: "gitlab", Label: "GitLab CI", Description: "Multi-stage pipeline in .gitlab-ci.yml"},
			{ID: 3, Key: "jenkins", Label: "Jenkins", Description: "Declarative Jenkinsfile pipeline"},
		}
		cChoice, _ := SelectOne(reader, w.in, w.out, "CI/CD automation provider", ciChoices, 0)
		ciProvider := cChoice.Key

		runnerChoices := []Choice{
			{ID: 1, Key: "cloud", Label: "Cloud / Hosted (Recommended)", Description: "Standard runners (ubuntu-latest / gitlab.com SaaS)"},
			{ID: 2, Key: "self-hosted", Label: "Existing self-hosted runner", Description: "Pre-configured external enterprise runner"},
			{ID: 3, Key: "in-cluster", Label: "In-cluster runner deployment", Description: "Scaffold dedicated runner pod in cluster"},
		}
		rChoice, _ := SelectOne(reader, w.in, w.out, "Runner environment", runnerChoices, 0)
		runnerType := rChoice.Key
		runnerTag := ""
		installRunner := false

		if runnerType == "self-hosted" {
			runnerTag = w.Ask(reader, "Runner tag/label", "self-hosted")
		} else if runnerType == "in-cluster" {
			runnerType = "self-hosted"
			installRunner = true
			runnerTag = "in-cluster-runner"
		}

		gitopsChoices := []Choice{
			{ID: 1, Key: "argocd", Label: "Argo CD (Recommended)", Description: "Kubernetes-native declarative GitOps"},
			{ID: 2, Key: "flux", Label: "Flux CD", Description: "GitOps toolkit controller for Kubernetes"},
			{ID: 3, Key: "none", Label: "None", Description: "Skip automated GitOps manifest"},
		}
		gChoice, _ := SelectOne(reader, w.in, w.out, "GitOps engine", gitopsChoices, 0)
		var gitopsCfg *config.GitOpsConfig
		if gChoice.Key != "none" {
			chartPath := fmt.Sprintf("charts/%s", name)
			if cfg.Modules.Kubernetes != nil && cfg.Modules.Kubernetes.Tool != "helm" {
				chartPath = "k8s"
			}
			gitopsCfg = &config.GitOpsConfig{
				Enabled:  true,
				Tool:     gChoice.Key,
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

	// Database Dependency
	dbChoices := []Choice{
		{ID: 1, Key: "none", Label: "None (Recommended)", Description: "Skip database integration"},
		{ID: 2, Key: "postgres", Label: "PostgreSQL", Description: "LTS postgres:16-alpine (relational DB)"},
		{ID: 3, Key: "mysql", Label: "MySQL", Description: "LTS mysql:8.0 (relational DB)"},
		{ID: 4, Key: "redis", Label: "Redis", Description: "redis:7-alpine (in-memory key-value/cache)"},
	}
	dbChoice, _ := SelectOne(reader, w.in, w.out, "Database dependency", dbChoices, 0)
	var dbCfg *config.DatabaseModuleConfig
	switch dbChoice.Key {
	case "postgres":
		topChoices := []Choice{
			{ID: 1, Key: "shared", Label: "Shared / Managed Cloud DB (Recommended)", Description: "Connect to external cloud DB without cluster bloat"},
			{ID: 2, Key: "dedicated", Label: "Dedicated In-Cluster (StatefulSet)", Description: "Deploy local StatefulSet with 10Gi PVC in namespace"},
		}
		topChoice, _ := SelectOne(reader, w.in, w.out, "Database cluster topology", topChoices, 0)
		migChoice := w.Ask(reader, "Include pre-upgrade Database Migration Job (Helm hook)? (yes/no)", "yes")
		enableMig := strings.ToLower(strings.TrimSpace(migChoice)) == "yes" || strings.ToLower(strings.TrimSpace(migChoice)) == "y"

		dbCfg = &config.DatabaseModuleConfig{
			Enabled:   true,
			Engine:    "postgres",
			Topology:  topChoice.Key,
			Migration: enableMig,
			Host:      "postgres.infra.svc.cluster.local",
			Port:      5432,
			Name:      name + "_db",
			User:      name + "_user",
		}
	case "mysql":
		topChoices := []Choice{
			{ID: 1, Key: "shared", Label: "Shared / Managed Cloud DB (Recommended)", Description: "Connect to external cloud DB without cluster bloat"},
			{ID: 2, Key: "dedicated", Label: "Dedicated In-Cluster (StatefulSet)", Description: "Deploy local StatefulSet with 10Gi PVC in namespace"},
		}
		topChoice, _ := SelectOne(reader, w.in, w.out, "Database cluster topology", topChoices, 0)
		migChoice := w.Ask(reader, "Include pre-upgrade Database Migration Job (Helm hook)? (yes/no)", "yes")
		enableMig := strings.ToLower(strings.TrimSpace(migChoice)) == "yes" || strings.ToLower(strings.TrimSpace(migChoice)) == "y"

		dbCfg = &config.DatabaseModuleConfig{
			Enabled:   true,
			Engine:    "mysql",
			Topology:  topChoice.Key,
			Migration: enableMig,
			Host:      "mysql.infra.svc.cluster.local",
			Port:      3306,
			Name:      name + "_db",
			User:      name + "_user",
		}
	case "redis":
		dbCfg = &config.DatabaseModuleConfig{
			Enabled:   true,
			Engine:    "redis",
			Topology:  "shared",
			Migration: false,
			Host:      "redis.infra.svc.cluster.local",
			Port:      6379,
		}
	}
	cfg.Modules.Database = dbCfg

	// Messaging Broker (RabbitMQ)
	msgChoices := []Choice{
		{ID: 1, Key: "no", Label: "No (Recommended)", Description: "Skip message broker"},
		{ID: 2, Key: "yes", Label: "Yes (RabbitMQ)", Description: "rabbitmq:3.13-management-alpine with Web UI & network policies"},
	}
	msgChoice, _ := SelectOne(reader, w.in, w.out, "Include RabbitMQ message broker?", msgChoices, 0)
	if msgChoice.Key == "yes" {
		cfg.Modules.Messaging = &config.MessagingModuleConfig{
			Enabled: true,
			Broker:  "rabbitmq",
			Host:    "rabbitmq.infra.svc.cluster.local",
			Port:    5672,
			VHost:   "/",
		}
	}

	// Final Summary Box
	fmt.Fprintln(w.out, "\n"+ansiGreen+ansiBold+"╭──────────────────────────────────────────────────────────╮"+ansiReset)
	fmt.Fprintln(w.out, ansiGreen+ansiBold+"│                ✨ Configuration Summary                  │"+ansiReset)
	fmt.Fprintln(w.out, ansiGreen+ansiBold+"├──────────────────────────────────────────────────────────┤"+ansiReset)
	fmt.Fprintf(w.out, "│  %-16s %-38s │\n", "Project:", name+" (v"+version+")")
	fmt.Fprintf(w.out, "│  %-16s %-38s │\n", "Environment:", envName+" (ns: "+namespace+")")
	if cfg.Modules.Kubernetes != nil && cfg.Modules.Kubernetes.Enabled {
		strategy := cfg.Modules.Kubernetes.Tool
		if cfg.Modules.Kubernetes.Autoscaling != nil && cfg.Modules.Kubernetes.Autoscaling.Enabled {
			strategy += fmt.Sprintf(" (replicas: %d-%d)", cfg.Modules.Kubernetes.Replicas, cfg.Modules.Kubernetes.Autoscaling.MaxReplicas)
		}
		if len(strategy) > 38 {
			strategy = strategy[:38]
		}
		fmt.Fprintf(w.out, "│  %-16s %-38s │\n", "Workload:", strategy)
	}
	if cfg.Modules.Kubernetes != nil && cfg.Modules.Kubernetes.Enabled && cfg.Modules.Kubernetes.Ingress {
		fmt.Fprintf(w.out, "│  %-16s %-38s │\n", "Ingress:", fmt.Sprintf("%s controller", cfg.Modules.Kubernetes.IngressClass))
	}
	if cfg.Modules.Database != nil && cfg.Modules.Database.Enabled {
		dbSummary := fmt.Sprintf("%s (%s)", cfg.Modules.Database.Engine, cfg.Modules.Database.Topology)
		if len(dbSummary) > 38 {
			dbSummary = dbSummary[:38]
		}
		fmt.Fprintf(w.out, "│  %-16s %-38s │\n", "Database:", dbSummary)
	}
	if cfg.Modules.Messaging != nil && cfg.Modules.Messaging.Enabled {
		fmt.Fprintf(w.out, "│  %-16s %-38s │\n", "Messaging:", "RabbitMQ (AMQP 5672)")
	}
	if cfg.Modules.CICD != nil && cfg.Modules.CICD.Enabled {
		ciInfo := cfg.Modules.CICD.Provider
		if cfg.Modules.CICD.GitOps != nil && cfg.Modules.CICD.GitOps.Enabled {
			ciInfo += " (" + cfg.Modules.CICD.GitOps.Tool + ")"
		}
		if len(ciInfo) > 38 {
			ciInfo = ciInfo[:38]
		}
		fmt.Fprintf(w.out, "│  %-16s %-38s │\n", "CI/CD & GitOps:", ciInfo)
	}
	fmt.Fprintln(w.out, ansiGreen+ansiBold+"╰──────────────────────────────────────────────────────────╯"+ansiReset)
	fmt.Fprintf(w.out, "\n%s✓ Configuration ready! Run '%smoreops plan%s' to preview changes.%s\n\n",
		ansiGreen, ansiBold, ansiGreen, ansiReset)

	return cfg, nil
}
