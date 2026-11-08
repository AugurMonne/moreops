package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const DefaultConfigFileName = "moreops.yaml"

// Config represents the declarative configuration for MoreOps.
type Config struct {
	Version      string                       `yaml:"version"`
	Project      ProjectMeta                  `yaml:"project"`
	Environments map[string]EnvironmentConfig `yaml:"environments"`
	Modules      ModulesConfig                `yaml:"modules"`
}

// ProjectMeta contains metadata about the application or infrastructure project.
type ProjectMeta struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version,omitempty"`
	Description string `yaml:"description,omitempty"`
	Language    string `yaml:"language,omitempty"`
	Framework   string `yaml:"framework,omitempty"`
}

// EnvironmentConfig defines the deployment target (k8s cluster, namespace, cloud).
type EnvironmentConfig struct {
	Cluster   string            `yaml:"cluster,omitempty"`   // Kubernetes context or cluster name
	Namespace string            `yaml:"namespace,omitempty"` // Kubernetes namespace
	Cloud     string            `yaml:"cloud,omitempty"`     // aws, gcp, azure, local, etc.
	Values    map[string]any    `yaml:"values,omitempty"`    // Environment-specific overrides
	Variables map[string]string `yaml:"variables,omitempty"` // Environment variables
}

// ModulesConfig holds configuration for all selectable modules.
type ModulesConfig struct {
	Containers *ContainersModuleConfig `yaml:"containers,omitempty"`
	Kubernetes *KubernetesModuleConfig `yaml:"kubernetes,omitempty"`
	Monitoring *MonitoringModuleConfig `yaml:"monitoring,omitempty"`
	Security   *SecurityModuleConfig   `yaml:"security,omitempty"`
	Storage    *StorageModuleConfig    `yaml:"storage,omitempty"`
	CICD       *CICDModuleConfig       `yaml:"cicd,omitempty"`
	Database   *DatabaseModuleConfig   `yaml:"database,omitempty"`
	Messaging  *MessagingModuleConfig  `yaml:"messaging,omitempty"`
}

// DatabaseModuleConfig configures database integration and topology.
type DatabaseModuleConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Engine    string `yaml:"engine,omitempty"`    // postgres, mysql, redis
	Topology  string `yaml:"topology,omitempty"`  // shared (managed external/cluster), dedicated (in-cluster)
	Migration bool   `yaml:"migration,omitempty"` // enable pre-upgrade migration Job hook
	Host      string `yaml:"host,omitempty"`
	Port      int    `yaml:"port,omitempty"`
	Name      string `yaml:"name,omitempty"`
	User      string `yaml:"user,omitempty"`
}

// MessagingModuleConfig configures message broker integration (RabbitMQ).
type MessagingModuleConfig struct {
	Enabled bool   `yaml:"enabled"`
	Broker  string `yaml:"broker,omitempty"` // rabbitmq
	Host    string `yaml:"host,omitempty"`
	Port    int    `yaml:"port,omitempty"`
	VHost   string `yaml:"vhost,omitempty"`
}

// ContainersModuleConfig configures container runtime and build setup.
type ContainersModuleConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Engine     string `yaml:"engine,omitempty"` // docker, containerd, podman
	Dockerfile string `yaml:"dockerfile,omitempty"`
	Image      string `yaml:"image,omitempty"`
	Compose    bool   `yaml:"compose,omitempty"`
}

// KubernetesModuleConfig configures Kubernetes workload deployment.
type KubernetesModuleConfig struct {
	Enabled      bool               `yaml:"enabled"`
	Tool         string             `yaml:"tool,omitempty"` // manifests, helm, kustomize
	Port         int                `yaml:"port,omitempty"`
	Replicas     int                `yaml:"replicas,omitempty"`
	Ingress      bool               `yaml:"ingress,omitempty"`
	IngressClass string             `yaml:"ingress_class,omitempty"` // nginx, traefik, caddy, haproxy, custom
	ConfigMap    bool               `yaml:"configmap,omitempty"`
	Autoscaling  *AutoscalingConfig `yaml:"autoscaling,omitempty"`
}

// AutoscalingConfig defines Horizontal Pod Autoscaler policies.
type AutoscalingConfig struct {
	Enabled       bool `yaml:"enabled"`
	MinReplicas   int  `yaml:"min_replicas,omitempty"`
	MaxReplicas   int  `yaml:"max_replicas,omitempty"`
	CPUPercentage int  `yaml:"cpu_percentage,omitempty"`
}

// MonitoringModuleConfig configures Prometheus, Grafana, Loki, Alertmanager, etc.
type MonitoringModuleConfig struct {
	Enabled        bool     `yaml:"enabled"`
	Prometheus     bool     `yaml:"prometheus,omitempty"`
	Grafana        bool     `yaml:"grafana,omitempty"`
	Alertmanager   bool     `yaml:"alertmanager,omitempty"`
	Loki           bool     `yaml:"loki,omitempty"`
	OpenTelemetry  bool     `yaml:"opentelemetry,omitempty"`
	MetricsPath    string   `yaml:"metrics_path,omitempty"`    // e.g. "/metrics"
	ScrapeInterval string   `yaml:"scrape_interval,omitempty"` // e.g. "15s"
	Dashboards     bool     `yaml:"dashboards,omitempty"`      // Deep Grafana Dashboards
	AlertRules     bool     `yaml:"alert_rules,omitempty"`     // PrometheusRule alerts
	Components     []string `yaml:"components,omitempty"`
}

// SecurityModuleConfig configures Vault, ESO, Trivy scanning, RBAC.
type SecurityModuleConfig struct {
	Enabled         bool                   `yaml:"enabled"`
	Vault           *VaultConfig           `yaml:"vault,omitempty"`
	ExternalSecrets *ExternalSecretsConfig `yaml:"external_secrets,omitempty"`
	RBAC            *RBACConfig            `yaml:"rbac,omitempty"`
	Trivy           *TrivyConfig           `yaml:"trivy,omitempty"`
	Components      []string               `yaml:"components,omitempty"`
}

// VaultConfig configures HashiCorp Vault integration.
type VaultConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Mode       string `yaml:"mode,omitempty"`        // "in-cluster-dev", "external"
	Address    string `yaml:"address,omitempty"`     // e.g. "http://vault:8200"
	AuthMethod string `yaml:"auth_method,omitempty"` // "kubernetes", "token", "approle"
}

// ExternalSecretsConfig configures External Secrets Operator (ESO).
type ExternalSecretsConfig struct {
	Enabled         bool     `yaml:"enabled"`
	RefreshInterval string   `yaml:"refresh_interval,omitempty"` // e.g. "1h"
	SecretStoreName string   `yaml:"secret_store_name,omitempty"`
	TargetSecrets   []string `yaml:"target_secrets,omitempty"`
}

// RBACConfig configures Role-Based Access Control generation.
type RBACConfig struct {
	Mode string `yaml:"mode"` // "strict", "manual", "disabled"
}

// TrivyConfig configures security scanning.
type TrivyConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Severity string `yaml:"severity,omitempty"` // "HIGH,CRITICAL"
}

// StorageModuleConfig configures S3/MinIO and backup (Velero, DB backups).
type StorageModuleConfig struct {
	Enabled        bool            `yaml:"enabled"`
	Provider       string          `yaml:"provider,omitempty"` // minio-dev, yandex, vkcloud, selectel, aws, custom
	Endpoint       string          `yaml:"endpoint,omitempty"`
	Region         string          `yaml:"region,omitempty"`
	Bucket         string          `yaml:"bucket,omitempty"`
	Velero         *VeleroConfig   `yaml:"velero,omitempty"`
	DatabaseBackup *DBBackupConfig `yaml:"database_backup,omitempty"`
	Components     []string        `yaml:"components,omitempty"`
}

// VeleroConfig configures cluster backup schedules using Velero.
type VeleroConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Schedule string `yaml:"schedule,omitempty"` // e.g. "0 2 * * *"
	TTL      string `yaml:"ttl,omitempty"`      // e.g. "720h"
}

// DBBackupConfig configures automated database backup CronJobs.
type DBBackupConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Engine    string `yaml:"engine,omitempty"`   // postgres, mysql
	Schedule  string `yaml:"schedule,omitempty"` // e.g. "0 3 * * *"
	Database  string `yaml:"database,omitempty"`
	SecretRef string `yaml:"secret_ref,omitempty"`
}

// CICDModuleConfig configures CI/CD, GitOps, IaC and Ansible automation.
type CICDModuleConfig struct {
	Enabled       bool          `yaml:"enabled"`
	Provider      string        `yaml:"provider,omitempty"`    // github, gitlab
	RunnerType    string        `yaml:"runner_type,omitempty"` // cloud, self-hosted
	RunnerTag     string        `yaml:"runner_tag,omitempty"`
	InstallRunner bool          `yaml:"install_runner,omitempty"`
	GitOps        *GitOpsConfig `yaml:"gitops,omitempty"`
	IaC           *IaCConfig    `yaml:"iac,omitempty"`
	Ansible       bool          `yaml:"ansible,omitempty"`
}

// GitOpsConfig configures Argo CD or Flux CD reconciliation.
type GitOpsConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Tool     string `yaml:"tool,omitempty"`     // argocd, flux
	RepoURL  string `yaml:"repo_url,omitempty"` // git repository URL
	Path     string `yaml:"path,omitempty"`     // path to chart or manifests
	AutoSync bool   `yaml:"auto_sync,omitempty"`
}

// IaCConfig configures OpenTofu / Terraform infrastructure as code.
type IaCConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Tool     string `yaml:"tool,omitempty"`     // tofu, terraform
	Provider string `yaml:"provider,omitempty"` // yandex, aws, selectel, local
}

// NewDefaultConfig creates a starter configuration with safe defaults.
func NewDefaultConfig(projectName string) *Config {
	if projectName == "" {
		projectName = "my-app"
	}
	return &Config{
		Version: "1",
		Project: ProjectMeta{
			Name:    projectName,
			Version: "0.1.0",
		},
		Environments: map[string]EnvironmentConfig{
			"dev": {
				Cluster:   "current-context",
				Namespace: "default",
				Cloud:     "local",
			},
			"prod": {
				Cluster:   "prod-cluster",
				Namespace: projectName + "-prod",
				Cloud:     "aws",
			},
		},
		Modules: ModulesConfig{
			Containers: &ContainersModuleConfig{
				Enabled: true,
				Engine:  "docker",
				Image:   projectName + ":latest",
			},
			Kubernetes: &KubernetesModuleConfig{
				Enabled:  true,
				Tool:     "manifests",
				Port:     8080,
				Replicas: 2,
			},
		},
	}
}

// LoadConfig loads and parses a moreops.yaml configuration file.
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigFileName
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse yaml in '%s': %w", path, err)
	}

	if cfg.Project.Name == "" {
		return nil, fmt.Errorf("invalid config: project.name is required")
	}

	return &cfg, nil
}

// SaveConfig saves the configuration to a file in YAML format.
func SaveConfig(path string, cfg *Config) error {
	if path == "" {
		path = DefaultConfigFileName
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory '%s': %w", dir, err)
		}
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file '%s': %w", path, err)
	}
	return nil
}

// GetEnvironment returns the environment configuration for the given name.
func (c *Config) GetEnvironment(envName string) (EnvironmentConfig, bool) {
	if c.Environments == nil {
		return EnvironmentConfig{}, false
	}
	env, ok := c.Environments[envName]
	return env, ok
}
