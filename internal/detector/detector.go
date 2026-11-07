package detector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// ProjectInfo contains detected project metadata and capabilities.
type ProjectInfo struct {
	Name          string
	Language      string
	Framework     string
	DefaultPort   int
	HasDockerfile bool
	HasCompose    bool
	HasK8s        bool
	HasHelm       bool
}

// Detector analyzes project files to detect languages and frameworks.
type Detector struct{}

// NewDetector creates a new Detector.
func NewDetector() *Detector {
	return &Detector{}
}

// Detect analyzes the target directory and returns detected project information.
func (d *Detector) Detect(dir string) (*ProjectInfo, error) {
	if dir == "" {
		dir = "."
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	info := &ProjectInfo{
		Name:        filepath.Base(absDir),
		DefaultPort: 8080,
	}

	// Detect container and k8s files
	info.HasDockerfile = fileExists(filepath.Join(dir, "Dockerfile"))
	info.HasCompose = fileExists(filepath.Join(dir, "docker-compose.yml")) ||
		fileExists(filepath.Join(dir, "docker-compose.yaml")) ||
		fileExists(filepath.Join(dir, "compose.yml")) ||
		fileExists(filepath.Join(dir, "compose.yaml"))

	info.HasK8s = dirExists(filepath.Join(dir, "k8s")) ||
		dirExists(filepath.Join(dir, "deploy")) ||
		dirExists(filepath.Join(dir, "manifests"))

	info.HasHelm = fileExists(filepath.Join(dir, "Chart.yaml")) ||
		dirExists(filepath.Join(dir, "charts")) ||
		dirExists(filepath.Join(dir, "helm"))

	// Detect language and framework
	info.Language = DetectLanguage(dir)
	if info.Language != "" {
		info.Framework = DetectFramework(dir, info.Language)
	}

	// Adjust default port based on language/framework
	switch info.Language {
	case "node":
		info.DefaultPort = 3000
		if pkgName := extractNodePackageName(dir); pkgName != "" {
			info.Name = pkgName
		}
	case "python":
		info.DefaultPort = 8000
	case "go":
		info.DefaultPort = 8080
		if modName := extractGoModuleName(dir); modName != "" {
			parts := strings.Split(modName, "/")
			info.Name = parts[len(parts)-1]
		}
	case "java":
		info.DefaultPort = 8080
	case "rust":
		info.DefaultPort = 8080
	}

	return info, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func extractNodePackageName(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return ""
	}
	var pkg struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return ""
	}
	return pkg.Name
}

func extractGoModuleName(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "module "))
		}
	}
	return ""
}
