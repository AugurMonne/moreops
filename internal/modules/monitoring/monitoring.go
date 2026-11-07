package monitoring

import (
	"context"
	"fmt"
	"path/filepath"

	"moreops/internal/core/config"
	"moreops/internal/core/diff"
	"moreops/internal/core/plugin"
)

// Module implements Prometheus, Grafana, Alertmanager, and deep observability management.
type Module struct{}

func init() {
	plugin.Register(&Module{})
}

func (m *Module) Name() string {
	return "monitoring"
}

func (m *Module) Description() string {
	return "Manages Prometheus Operator, Grafana dashboards, Alertmanager, and deep cluster/pod metrics"
}

func (m *Module) Plan(ctx context.Context, cfg *config.Config, envName string) (*plugin.PlanResult, error) {
	if cfg.Modules.Monitoring == nil || !cfg.Modules.Monitoring.Enabled {
		return nil, nil
	}

	env, _ := cfg.GetEnvironment(envName)
	namespace := env.Namespace
	if namespace == "" {
		namespace = "default"
	}

	appName := cfg.Project.Name
	monCfg := cfg.Modules.Monitoring

	metricsPath := monCfg.MetricsPath
	if metricsPath == "" {
		metricsPath = "/metrics"
	}
	scrapeInterval := monCfg.ScrapeInterval
	if scrapeInterval == "" {
		scrapeInterval = "15s"
	}

	res := &plugin.PlanResult{
		ModuleName:  m.Name(),
		FileChanges: make([]diff.FileChange, 0),
		Summary:     fmt.Sprintf("Monitoring and deep observability configuration for %s", appName),
	}

	isHelm := cfg.Modules.Kubernetes != nil && cfg.Modules.Kubernetes.Tool == "helm"
	var outDir string
	if isHelm {
		outDir = filepath.Join("charts", appName, "templates", "monitoring")
	} else {
		outDir = filepath.Join("k8s", "monitoring")
	}

	files := make(map[string]string)

	// 1. Prometheus Operator ServiceMonitor (Deep Scraping)
	serviceMonitorFile := filepath.Join(outDir, "servicemonitor.yaml")
	files[serviceMonitorFile] = fmt.Sprintf(`apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: %s
  namespace: %s
  labels:
    app.kubernetes.io/name: %s
    release: prometheus-stack
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: %s
  endpoints:
    - port: http
      path: %s
      interval: %s
      scrapeTimeout: 10s
      relabelings:
        - action: replace
          sourceLabels: [__meta_kubernetes_pod_node_name]
          targetLabel: node
        - action: replace
          sourceLabels: [__meta_kubernetes_pod_name]
          targetLabel: pod
        - action: replace
          sourceLabels: [__meta_kubernetes_namespace]
          targetLabel: namespace
`, appName, namespace, appName, appName, metricsPath, scrapeInterval)

	// 2. PrometheusRule (Deep Alerting: Latency p99, Error Rates, Node & Pod Resource Exhaustion)
	if monCfg.AlertRules {
		rulesFile := filepath.Join(outDir, "prometheusrule.yaml")
		files[rulesFile] = fmt.Sprintf(`apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: %s-deep-alerts
  namespace: %s
  labels:
    app.kubernetes.io/name: %s
    role: alert-rules
spec:
  groups:
    - name: %s-performance-and-health
      rules:
        # 1. High Inter-Pod / HTTP Latency Alert
        - alert: HighRequestDurationP99
          expr: histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket{app_kubernetes_io_name="%s"}[5m])) by (le, pod, node)) > 1.0
          for: 2m
          labels:
            severity: warning
          annotations:
            summary: "High p99 request duration detected on pod {{ $labels.pod }} (node: {{ $labels.node }})"
            description: "99%% of requests to {{ $labels.pod }} take longer than 1s (current value: {{ $value }}s)."

        # 2. High 5xx Error Rate Alert
        - alert: High5xxErrorRate
          expr: (sum(rate(http_requests_total{app_kubernetes_io_name="%s",status=~"5.."}[5m])) / sum(rate(http_requests_total{app_kubernetes_io_name="%s"}[5m]))) * 100 > 5
          for: 3m
          labels:
            severity: critical
          annotations:
            summary: "5xx error rate exceeded 5%% for %s"
            description: "Application is returning 5xx responses for more than 5%% of total requests."

        # 3. Pod CPU Consumption Near Limit (Top Pod Load)
        - alert: PodHighCPULoad
          expr: (sum(rate(container_cpu_usage_seconds_total{container!="POD",container!="",pod=~"%s-.*"}[5m])) by (pod, node) / sum(kube_pod_container_resource_limits{resource="cpu",pod=~"%s-.*"}) by (pod, node)) * 100 > 85
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "Pod {{ $labels.pod }} on node {{ $labels.node }} is consuming over 85%% CPU limit"
            description: "Pod CPU load is at {{ $value }}%% of its allocated limit."

        # 4. Pod CrashLooping / Frequent Restarts
        - alert: PodFrequentlyRestarting
          expr: increase(kube_pod_container_status_restarts_total{pod=~"%s-.*"}[15m]) > 2
          labels:
            severity: critical
          annotations:
            summary: "Pod {{ $labels.pod }} is restarting frequently"
            description: "Pod restarted {{ $value }} times within the last 15 minutes."
`, appName, namespace, appName, appName, appName, appName, appName, appName, appName, appName, appName)
	}

	// 3. Deep Grafana Dashboard (ConfigMap auto-loaded by Grafana sidecar)
	if monCfg.Dashboards {
		dashFile := filepath.Join(outDir, "grafana-dashboard-deep.yaml")
		files[dashFile] = fmt.Sprintf(`apiVersion: v1
kind: ConfigMap
metadata:
  name: %s-deep-observability-dashboard
  namespace: %s
  labels:
    grafana_dashboard: "1"
    app.kubernetes.io/name: %s
data:
  %s-observability.json: |-
    {
      "annotations": { "list": [] },
      "editable": true,
      "title": "%s - Deep Performance & Cluster Observability",
      "tags": ["moreops", "kubernetes", "%s"],
      "timezone": "browser",
      "schemaVersion": 36,
      "refresh": "10s",
      "panels": [
        {
          "id": 1,
          "title": "Top Pods by CPU Utilization",
          "type": "timeseries",
          "gridPos": { "h": 8, "w": 12, "x": 0, "y": 0 },
          "targets": [
            {
              "expr": "topk(5, sum(rate(container_cpu_usage_seconds_total{namespace=\"%s\"}[5m])) by (pod, node))",
              "legendFormat": "{{pod}} (on {{node}})"
            }
          ]
        },
        {
          "id": 2,
          "title": "Top Pods by Memory Working Set",
          "type": "timeseries",
          "gridPos": { "h": 8, "w": 12, "x": 12, "y": 0 },
          "targets": [
            {
              "expr": "topk(5, sum(container_memory_working_set_bytes{namespace=\"%s\"}) by (pod, node))",
              "legendFormat": "{{pod}} (on {{node}})"
            }
          ]
        },
        {
          "id": 3,
          "title": "Inter-Pod & Service Request Latency Duration (p50 / p95 / p99)",
          "type": "timeseries",
          "gridPos": { "h": 8, "w": 12, "x": 0, "y": 8 },
          "targets": [
            {
              "expr": "histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket{namespace=\"%s\"}[5m])) by (le))",
              "legendFormat": "p99 Duration (seconds)"
            },
            {
              "expr": "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{namespace=\"%s\"}[5m])) by (le))",
              "legendFormat": "p95 Duration (seconds)"
            },
            {
              "expr": "histogram_quantile(0.50, sum(rate(http_request_duration_seconds_bucket{namespace=\"%s\"}[5m])) by (le))",
              "legendFormat": "p50 Duration (seconds)"
            }
          ]
        },
        {
          "id": 4,
          "title": "Network Transmit / Receive Rate between Pods",
          "type": "timeseries",
          "gridPos": { "h": 8, "w": 12, "x": 12, "y": 8 },
          "targets": [
            {
              "expr": "sum(rate(container_network_receive_bytes_total{namespace=\"%s\"}[5m])) by (pod)",
              "legendFormat": "RX: {{pod}}"
            },
            {
              "expr": "sum(rate(container_network_transmit_bytes_total{namespace=\"%s\"}[5m])) by (pod)",
              "legendFormat": "TX: {{pod}}"
            }
          ]
        },
        {
          "id": 5,
          "title": "Request Throughput (RPS) & Error Rate",
          "type": "timeseries",
          "gridPos": { "h": 7, "w": 12, "x": 0, "y": 16 },
          "targets": [
            {
              "expr": "sum(rate(http_requests_total{namespace=\"%s\"}[5m])) by (status)",
              "legendFormat": "HTTP Status {{status}}"
            }
          ]
        },
        {
          "id": 6,
          "title": "Pod Restarts & Failures",
          "type": "timeseries",
          "gridPos": { "h": 7, "w": 12, "x": 12, "y": 16 },
          "targets": [
            {
              "expr": "sum(increase(kube_pod_container_status_restarts_total{namespace=\"%s\"}[1h])) by (pod)",
              "legendFormat": "Restarts (last 1h): {{pod}}"
            }
          ]
        }
      ]
    }
`, appName, namespace, appName, appName, appName, appName, namespace, namespace, namespace, namespace, namespace, namespace, namespace, namespace, namespace)
	}

	// 4. Kube-Prometheus-Stack Recommended Values file
	valuesStackFile := filepath.Join("monitoring", "kube-prometheus-stack-values.yaml")
	files[valuesStackFile] = `prometheus:
  prometheusSpec:
    scrapeInterval: 15s
    evaluationInterval: 15s
    serviceMonitorSelectorNilUsesHelmValues: false
    podMonitorSelectorNilUsesHelmValues: false
    ruleSelectorNilUsesHelmValues: false

grafana:
  sidecar:
    dashboards:
      enabled: true
      label: grafana_dashboard
      labelValue: "1"
      searchNamespace: ALL

nodeExporter:
  enabled: true

kubeStateMetrics:
  enabled: true

alertmanager:
  enabled: true
`

	// 5. If Loki or OpenTelemetry selected, add collector configs
	if monCfg.Loki {
		lokiFile := filepath.Join(outDir, "loki-promtail-values.yaml")
		files[lokiFile] = `promtail:
  enabled: true
  config:
    snippets:
      extraScrapeConfigs: |
        - job_name: app-logs
          kubernetes_sd_configs:
            - role: pod
          relabel_configs:
            - source_labels: [__meta_kubernetes_pod_name]
              target_label: pod
            - source_labels: [__meta_kubernetes_namespace]
              target_label: namespace
`
	}

	if monCfg.OpenTelemetry {
		otelFile := filepath.Join(outDir, "otel-collector.yaml")
		files[otelFile] = fmt.Sprintf(`apiVersion: v1
kind: ConfigMap
metadata:
  name: otel-collector-config
  namespace: %s
data:
  config.yaml: |
    receivers:
      otlp:
        protocols:
          grpc:
            endpoint: 0.0.0.0:4317
          http:
            endpoint: 0.0.0.0:4318
    processors:
      batch:
        timeout: 1s
    exporters:
      logging:
        loglevel: info
    service:
      pipelines:
        traces:
          receivers: [otlp]
          processors: [batch]
          exporters: [logging]
`, namespace)
	}

	for path, content := range files {
		fc, err := diff.ComputeFileChange(path, content)
		if err != nil {
			return nil, err
		}
		res.FileChanges = append(res.FileChanges, fc)
	}

	return res, nil
}

func (m *Module) Apply(ctx context.Context, plan *plugin.PlanResult) error {
	return plugin.ExecutePlan(plan)
}
