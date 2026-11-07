package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// IssueSeverity indicates error or warning level.
type IssueSeverity string

const (
	SeverityError   IssueSeverity = "ERROR"
	SeverityWarning IssueSeverity = "WARNING"
)

// Issue represents a security, reliability, or consistency finding.
type Issue struct {
	RuleID   string        `json:"rule_id"`
	Severity IssueSeverity `json:"severity"`
	Resource string        `json:"resource"`
	File     string        `json:"file"`
	Message  string        `json:"message"`
}

// Validator executes static analysis on manifests.
type Validator struct{}

// NewValidator creates a Validator.
func NewValidator() *Validator {
	return &Validator{}
}

// RawK8sResource minimal unmarshaling structure.
type RawK8sResource struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string            `yaml:"name"`
		Namespace string            `yaml:"namespace"`
		Labels    map[string]string `yaml:"labels"`
	} `yaml:"metadata"`
	Spec map[string]any `yaml:"spec"`
}

// ValidateDirectory inspects all YAML manifests in a directory tree.
func (v *Validator) ValidateDirectory(root string) ([]Issue, error) {
	var issues []Issue
	var resources []RawK8sResource
	var resourceFiles []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}
		// Skip helm helper templates which aren't valid single YAML
		if strings.HasSuffix(path, "_helpers.tpl") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		// Split multi-document YAML
		docs := strings.Split(string(data), "\n---")
		for _, doc := range docs {
			trimmed := strings.TrimSpace(doc)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}

			// Check YAML syntax
			var node yaml.Node
			if err := yaml.Unmarshal([]byte(trimmed), &node); err != nil {
				// If it's a Helm template with Go template brackets {{ }}, ignore strict syntax
				if strings.Contains(trimmed, "{{") && strings.Contains(trimmed, "}}") {
					continue
				}
				issues = append(issues, Issue{
					RuleID:   "SYN001",
					Severity: SeverityError,
					File:     path,
					Message:  fmt.Sprintf("Invalid YAML syntax: %v", err),
				})
				continue
			}

			var res RawK8sResource
			if err := yaml.Unmarshal([]byte(trimmed), &res); err == nil && res.Kind != "" {
				resources = append(resources, res)
				resourceFiles = append(resourceFiles, path)
				issues = append(issues, checkSecurityRules(res, path)...)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Consistency cross-checks
	issues = append(issues, checkConsistency(resources, resourceFiles)...)

	return issues, nil
}

func checkSecurityRules(res RawK8sResource, file string) []Issue {
	var issues []Issue
	resIdentifier := fmt.Sprintf("%s/%s", res.Kind, res.Metadata.Name)

	if res.Kind == "Deployment" || res.Kind == "StatefulSet" || res.Kind == "DaemonSet" {
		template, ok := res.Spec["template"].(map[string]any)
		if !ok {
			return issues
		}
		spec, ok := template["spec"].(map[string]any)
		if !ok {
			return issues
		}

		// Check Pod-level securityContext
		podSecCtx, _ := spec["securityContext"].(map[string]any)
		runAsNonRoot := false
		if podSecCtx != nil {
			if nonRoot, ok := podSecCtx["runAsNonRoot"].(bool); ok && nonRoot {
				runAsNonRoot = true
			}
		}

		containers, _ := spec["containers"].([]any)
		for idx, c := range containers {
			cMap, ok := c.(map[string]any)
			if !ok {
				continue
			}
			cName, _ := cMap["name"].(string)
			if cName == "" {
				cName = fmt.Sprintf("container-%d", idx)
			}

			cSecCtx, _ := cMap["securityContext"].(map[string]any)
			containerNonRoot := runAsNonRoot
			if cSecCtx != nil {
				if nr, ok := cSecCtx["runAsNonRoot"].(bool); ok && nr {
					containerNonRoot = true
				}
			}

			// SEC001: runAsNonRoot
			if !containerNonRoot {
				issues = append(issues, Issue{
					RuleID:   "SEC001",
					Severity: SeverityWarning,
					Resource: resIdentifier,
					File:     file,
					Message:  fmt.Sprintf("Container '%s' should enforce runAsNonRoot: true", cName),
				})
			}

			// SEC002: allowPrivilegeEscalation
			if cSecCtx == nil || cSecCtx["allowPrivilegeEscalation"] != false {
				issues = append(issues, Issue{
					RuleID:   "SEC002",
					Severity: SeverityError,
					Resource: resIdentifier,
					File:     file,
					Message:  fmt.Sprintf("Container '%s' must explicitly set allowPrivilegeEscalation: false", cName),
				})
			}

			// SEC003: readOnlyRootFilesystem
			if cSecCtx == nil || cSecCtx["readOnlyRootFilesystem"] != true {
				issues = append(issues, Issue{
					RuleID:   "SEC003",
					Severity: SeverityWarning,
					Resource: resIdentifier,
					File:     file,
					Message:  fmt.Sprintf("Container '%s' should set readOnlyRootFilesystem: true", cName),
				})
			}

			// SEC004: resources
			resMap, _ := cMap["resources"].(map[string]any)
			if resMap == nil {
				issues = append(issues, Issue{
					RuleID:   "SEC004",
					Severity: SeverityWarning,
					Resource: resIdentifier,
					File:     file,
					Message:  fmt.Sprintf("Container '%s' has no CPU/Memory requests and limits specified", cName),
				})
			}

			// REL001: liveness and readiness probes
			if cMap["livenessProbe"] == nil || cMap["readinessProbe"] == nil {
				issues = append(issues, Issue{
					RuleID:   "REL001",
					Severity: SeverityWarning,
					Resource: resIdentifier,
					File:     file,
					Message:  fmt.Sprintf("Container '%s' missing livenessProbe or readinessProbe", cName),
				})
			}
		}
	}

	return issues
}

func checkConsistency(resources []RawK8sResource, files []string) []Issue {
	var issues []Issue

	deployments := make(map[string]RawK8sResource)
	services := make(map[string]RawK8sResource)
	fileMap := make(map[string]string)

	for i, r := range resources {
		fileMap[r.Kind+"/"+r.Metadata.Name] = files[i]
		if r.Kind == "Deployment" {
			deployments[r.Metadata.Name] = r
		} else if r.Kind == "Service" {
			services[r.Metadata.Name] = r
		}
	}

	// CON001: Service selector matching Deployment labels
	for svcName, svc := range services {
		selector, ok := svc.Spec["selector"].(map[string]any)
		if !ok || len(selector) == 0 {
			continue
		}

		matchedAny := false
		for _, dep := range deployments {
			tmpl, ok := dep.Spec["template"].(map[string]any)
			if !ok {
				continue
			}
			meta, ok := tmpl["metadata"].(map[string]any)
			if !ok {
				continue
			}
			labels, ok := meta["labels"].(map[string]any)
			if !ok {
				continue
			}

			allMatch := true
			for k, v := range selector {
				if labels[k] != v {
					allMatch = false
					break
				}
			}
			if allMatch {
				matchedAny = true
				break
			}
		}

		if len(deployments) > 0 && !matchedAny {
			issues = append(issues, Issue{
				RuleID:   "CON001",
				Severity: SeverityError,
				Resource: "Service/" + svcName,
				File:     fileMap["Service/"+svcName],
				Message:  "Service selector does not match any Deployment template labels",
			})
		}
	}

	return issues
}
