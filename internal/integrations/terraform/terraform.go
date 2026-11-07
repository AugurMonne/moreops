package terraform

import "os/exec"

// FindCLI looks for tofu or terraform in PATH.
func FindCLI() (string, error) {
	if p, err := exec.LookPath("tofu"); err == nil {
		return p, nil
	}
	return exec.LookPath("terraform")
}
