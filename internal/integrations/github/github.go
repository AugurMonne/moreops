package github

import "fmt"

// FormatRunner returns runs-on expression for GitHub Actions.
func FormatRunner(runnerType, runnerTag string) string {
	if runnerType == "self-hosted" {
		if runnerTag == "" {
			runnerTag = "self-hosted"
		}
		return fmt.Sprintf("[%s]", runnerTag)
	}
	return "ubuntu-latest"
}
