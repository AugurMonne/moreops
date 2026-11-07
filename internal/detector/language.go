package detector

import (
	"path/filepath"
)

// DetectLanguage inspects files in the directory to determine the primary programming language.
func DetectLanguage(dir string) string {
	checks := []struct {
		file string
		lang string
	}{
		{"go.mod", "go"},
		{"package.json", "node"},
		{"requirements.txt", "python"},
		{"pyproject.toml", "python"},
		{"Pipfile", "python"},
		{"Cargo.toml", "rust"},
		{"pom.xml", "java"},
		{"build.gradle", "java"},
		{"Gemfile", "ruby"},
	}

	for _, check := range checks {
		if fileExists(filepath.Join(dir, check.file)) {
			return check.lang
		}
	}

	return ""
}
