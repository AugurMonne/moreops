package monitoring

import (
	"context"
	"strings"
	"testing"

	"moreops/internal/core/config"
)

func TestMonitoringModulePlan(t *testing.T) {
	cfg := config.NewDefaultConfig("analytics-srv")
	cfg.Modules.Kubernetes = &config.KubernetesModuleConfig{
		Enabled: true,
		Tool:    "helm",
	}
	cfg.Modules.Monitoring = &config.MonitoringModuleConfig{
		Enabled:        true,
		Prometheus:     true,
		Grafana:        true,
		Alertmanager:   true,
		MetricsPath:    "/metrics",
		ScrapeInterval: "15s",
		Dashboards:     true,
		AlertRules:     true,
	}

	mod := &Module{}
	plan, err := mod.Plan(context.Background(), cfg, "dev")
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}

	if plan == nil {
		t.Fatalf("expected non-nil plan")
	}

	foundServiceMonitor := false
	foundPrometheusRule := false
	foundDashboard := false
	foundKubePromStackValues := false

	for _, fc := range plan.FileChanges {
		if strings.Contains(fc.Path, "servicemonitor.yaml") {
			foundServiceMonitor = true
			if !strings.Contains(fc.NewData, "targetLabel: node") || !strings.Contains(fc.NewData, "targetLabel: pod") {
				t.Errorf("expected node and pod relabeling in ServiceMonitor")
			}
		}
		if strings.Contains(fc.Path, "prometheusrule.yaml") {
			foundPrometheusRule = true
			if !strings.Contains(fc.NewData, "HighRequestDurationP99") {
				t.Errorf("expected HighRequestDurationP99 alert rule")
			}
		}
		if strings.Contains(fc.Path, "grafana-dashboard-deep.yaml") {
			foundDashboard = true
			if !strings.Contains(fc.NewData, "Top Pods by CPU Utilization") || !strings.Contains(fc.NewData, "Inter-Pod & Service Request Latency Duration") {
				t.Errorf("expected deep metrics panels in Grafana dashboard")
			}
		}
		if strings.Contains(fc.Path, "kube-prometheus-stack-values.yaml") {
			foundKubePromStackValues = true
		}
	}

	if !foundServiceMonitor {
		t.Errorf("missing servicemonitor.yaml")
	}
	if !foundPrometheusRule {
		t.Errorf("missing prometheusrule.yaml")
	}
	if !foundDashboard {
		t.Errorf("missing grafana-dashboard-deep.yaml")
	}
	if !foundKubePromStackValues {
		t.Errorf("missing kube-prometheus-stack-values.yaml")
	}
}

func TestMonitoringModuleDisabled(t *testing.T) {
	cfg := config.NewDefaultConfig("app")
	cfg.Modules.Monitoring = nil

	mod := &Module{}
	plan, err := mod.Plan(context.Background(), cfg, "dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan != nil {
		t.Errorf("expected nil plan when monitoring is disabled")
	}
}
