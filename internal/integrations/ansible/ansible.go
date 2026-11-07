package ansible

import "os/exec"

// Available checks if ansible is installed.
func Available() bool {
	_, err := exec.LookPath("ansible-playbook")
	return err == nil
}
