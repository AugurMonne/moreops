package gitlab

import "fmt"

// FormatTags returns YAML tags block for GitLab CI.
func FormatTags(runnerType, runnerTag string) string {
	if runnerType == "self-hosted" && runnerTag != "" {
		return fmt.Sprintf("\n  tags:\n    - %s", runnerTag)
	}
	return ""
}
