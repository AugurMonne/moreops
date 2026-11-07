package trivy

import "os/exec"

// Available checks if Trivy CLI is locally installed.
func Available() bool {
	_, err := exec.LookPath("trivy")
	return err == nil
}

// DockerFallbackImage returns the official container image for running Trivy without installation.
func DockerFallbackImage() string {
	return "aquasec/trivy:latest"
}
