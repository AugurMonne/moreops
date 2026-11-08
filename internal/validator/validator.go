package validator

import (
	"bufio"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
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

// ValidateSourceFiles checks the syntax of application code (e.g. Go source files).
func (v *Validator) ValidateSourceFiles(root string) ([]Issue, error) {
	var issues []Issue
	fset := token.NewFileSet()

	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		if info.IsDir() {
			if path == root || path == "." || info.Name() == "." {
				return nil
			}
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" || name == "charts" || name == "k8s" {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Ext(path) == ".go" {
			_, parseErr := parser.ParseFile(fset, path, nil, parser.AllErrors)
			if parseErr != nil {
				issues = append(issues, Issue{
					RuleID:   "SYN002",
					Severity: SeverityError,
					Resource: "SourceCode/" + filepath.Base(path),
					File:     path,
					Message:  fmt.Sprintf("Language syntax error: %v", parseErr),
				})
			}
		}
		return nil
	})

	return issues, nil
}

func sanitizeHelmYAML(content string) string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(content))
	re := regexp.MustCompile(`\{\{[^}]*\}\}`)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Check if line is a toYaml or include block under a parent key
		if strings.HasPrefix(trimmed, "{{- toYaml") || strings.HasPrefix(trimmed, "{{ toYaml") ||
			strings.HasPrefix(trimmed, "{{- include") || strings.HasPrefix(trimmed, "{{ include") {
			indent := len(line) - len(strings.TrimLeft(line, " "))
			lines = append(lines, strings.Repeat(" ", indent)+"placeholder: val")
			continue
		}

		// Full-line control structure (e.g. {{- if ... }}, {{- end }}, {{- with ... }})
		if (strings.HasPrefix(trimmed, "{{-") || strings.HasPrefix(trimmed, "{{")) &&
			(strings.HasSuffix(trimmed, "-}}") || strings.HasSuffix(trimmed, "}}")) {
			colonIdx := strings.Index(line, ":")
			firstTpl := strings.Index(line, "{{")
			if colonIdx == -1 || colonIdx > firstTpl {
				indent := len(line) - len(strings.TrimLeft(line, " "))
				lines = append(lines, strings.Repeat(" ", indent)+"# helm-tpl")
				continue
			}
		}

		// Replace inline {{ ... }} with a bare identifier helm_val (not quoted, so helm_val-suffix works)
		cleaned := re.ReplaceAllString(line, "helm_val")
		lines = append(lines, cleaned)
	}
	return strings.Join(lines, "\n")
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
				sanitized := sanitizeHelmYAML(trimmed)
				if errSanitized := yaml.Unmarshal([]byte(sanitized), &node); errSanitized != nil {
					issues = append(issues, Issue{
						RuleID:   "SYN001",
						Severity: SeverityError,
						Resource: "Manifest/" + filepath.Base(path),
						File:     path,
						Message:  fmt.Sprintf("Invalid YAML syntax: %v", errSanitized),
					})
					continue
				}
			}

			var res RawK8sResource
			if err := yaml.Unmarshal([]byte(trimmed), &res); err != nil {
				sanitized := sanitizeHelmYAML(trimmed)
				_ = yaml.Unmarshal([]byte(sanitized), &res)
			}

			if res.Kind != "" {
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

func findValuesYaml(file string) map[string]any {
	dir := filepath.Dir(file)
	for i := 0; i < 3; i++ {
		valPath := filepath.Join(dir, "values.yaml")
		if data, err := os.ReadFile(valPath); err == nil {
			var m map[string]any
			if err := yaml.Unmarshal(data, &m); err == nil {
				return m
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return nil
}

func checkSecurityRules(res RawK8sResource, file string) []Issue {
	var issues []Issue
	resIdentifier := fmt.Sprintf("%s/%s", res.Kind, res.Metadata.Name)
	valuesMap := findValuesYaml(file)

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
		if podSecCtx == nil && valuesMap != nil {
			if v, ok := valuesMap["podSecurityContext"].(map[string]any); ok {
				podSecCtx = v
			}
		} else if podSecCtx != nil && valuesMap != nil {
			if v, ok := valuesMap["podSecurityContext"].(map[string]any); ok {
				for k, val := range v {
					if _, exists := podSecCtx[k]; !exists {
						podSecCtx[k] = val
					}
				}
			}
		}

		runAsNonRoot := false
		if podSecCtx != nil {
			if nonRoot, ok := podSecCtx["runAsNonRoot"].(bool); ok && nonRoot {
				runAsNonRoot = true
			}
			if runAsUser, ok := podSecCtx["runAsUser"].(int); ok && runAsUser == 0 {
				issues = append(issues, Issue{
					RuleID:   "SEC008",
					Severity: SeverityError,
					Resource: resIdentifier,
					File:     file,
					Message:  "Pod securityContext explicitly sets runAsUser: 0 (root)",
				})
			}
		}

		// SEC006: Host namespace isolation (hostNetwork, hostPID, hostIPC)
		if hostNet, ok := spec["hostNetwork"].(bool); ok && hostNet {
			issues = append(issues, Issue{
				RuleID:   "SEC006",
				Severity: SeverityError,
				Resource: resIdentifier,
				File:     file,
				Message:  "Pod spec must not set hostNetwork: true",
			})
		}
		if hostPID, ok := spec["hostPID"].(bool); ok && hostPID {
			issues = append(issues, Issue{
				RuleID:   "SEC006",
				Severity: SeverityError,
				Resource: resIdentifier,
				File:     file,
				Message:  "Pod spec must not set hostPID: true",
			})
		}
		if hostIPC, ok := spec["hostIPC"].(bool); ok && hostIPC {
			issues = append(issues, Issue{
				RuleID:   "SEC006",
				Severity: SeverityError,
				Resource: resIdentifier,
				File:     file,
				Message:  "Pod spec must not set hostIPC: true",
			})
		}

		// SEC007: hostPath volume forbidden
		if volumes, ok := spec["volumes"].([]any); ok {
			for _, v := range volumes {
				if vMap, ok := v.(map[string]any); ok {
					if _, hasHostPath := vMap["hostPath"]; hasHostPath {
						vName, _ := vMap["name"].(string)
						issues = append(issues, Issue{
							RuleID:   "SEC007",
							Severity: SeverityError,
							Resource: resIdentifier,
							File:     file,
							Message:  fmt.Sprintf("Volume '%s' uses insecure hostPath volume mount", vName),
						})
					}
				}
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
			if cSecCtx == nil && valuesMap != nil {
				if v, ok := valuesMap["securityContext"].(map[string]any); ok {
					cSecCtx = v
				}
			} else if cSecCtx != nil && valuesMap != nil {
				if v, ok := valuesMap["securityContext"].(map[string]any); ok {
					for k, val := range v {
						if _, exists := cSecCtx[k]; !exists {
							cSecCtx[k] = val
						}
					}
				}
			}

			if cMap["resources"] == nil && valuesMap != nil {
				if v, ok := valuesMap["resources"].(map[string]any); ok {
					cMap["resources"] = v
				}
			} else if cMap["resources"] != nil && valuesMap != nil {
				if resMap, ok := cMap["resources"].(map[string]any); ok {
					if v, ok := valuesMap["resources"].(map[string]any); ok {
						for k, val := range v {
							if _, exists := resMap[k]; !exists {
								resMap[k] = val
							}
						}
					}
				}
			}

			containerNonRoot := runAsNonRoot
			if cSecCtx != nil {
				if nr, ok := cSecCtx["runAsNonRoot"].(bool); ok && nr {
					containerNonRoot = true
				}
			}

			// SEC005: privileged container execution forbidden
			if cSecCtx != nil {
				if priv, ok := cSecCtx["privileged"].(bool); ok && priv {
					issues = append(issues, Issue{
						RuleID:   "SEC005",
						Severity: SeverityError,
						Resource: resIdentifier,
						File:     file,
						Message:  fmt.Sprintf("Container '%s' must not run in privileged mode (privileged: true)", cName),
					})
				}
				if runAsUser, ok := cSecCtx["runAsUser"].(int); ok && runAsUser == 0 {
					issues = append(issues, Issue{
						RuleID:   "SEC008",
						Severity: SeverityError,
						Resource: resIdentifier,
						File:     file,
						Message:  fmt.Sprintf("Container '%s' explicitly sets runAsUser: 0 (root)", cName),
					})
				}
			}

			// SEC009: Drop capabilities check
			hasDropAll := false
			if cSecCtx != nil {
				if caps, ok := cSecCtx["capabilities"].(map[string]any); ok {
					if dropList, ok := caps["drop"].([]any); ok {
						for _, d := range dropList {
							if dStr, ok := d.(string); ok && strings.ToUpper(dStr) == "ALL" {
								hasDropAll = true
								break
							}
						}
					}
				}
			}
			if !hasDropAll {
				issues = append(issues, Issue{
					RuleID:   "SEC009",
					Severity: SeverityWarning,
					Resource: resIdentifier,
					File:     file,
					Message:  fmt.Sprintf("Container '%s' should drop all capabilities (capabilities.drop: [\"ALL\"])", cName),
				})
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

	workloads := make(map[string]RawK8sResource)
	services := make(map[string]RawK8sResource)
	fileMap := make(map[string]string)

	for i, r := range resources {
		fileMap[r.Kind+"/"+r.Metadata.Name] = files[i]
		if r.Kind == "Deployment" || r.Kind == "StatefulSet" || r.Kind == "DaemonSet" {
			workloads[r.Metadata.Name] = r
		} else if r.Kind == "Service" {
			services[r.Metadata.Name] = r
		}
	}

	// CON001: Service selector matching workload template labels
	for svcName, svc := range services {
		selector, ok := svc.Spec["selector"].(map[string]any)
		if !ok || len(selector) == 0 {
			continue
		}

		matchedAny := false
		for _, w := range workloads {
			tmpl, ok := w.Spec["template"].(map[string]any)
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

		if len(workloads) > 0 && !matchedAny {
			issues = append(issues, Issue{
				RuleID:   "CON001",
				Severity: SeverityError,
				Resource: "Service/" + svcName,
				File:     fileMap["Service/"+svcName],
				Message:  "Service selector does not match any Deployment or StatefulSet template labels",
			})
		}
	}

	return issues
}
