package iac

import (
	"context"
	"fmt"
	"os/exec"
)

type Runner struct {
	Tool string // tofu or terraform
}

func NewRunner(tool string) *Runner {
	if tool == "" {
		tool = "tofu"
	}
	return &Runner{Tool: tool}
}

func (r *Runner) Run(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, r.Tool, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s failed: %w (output: %s)", r.Tool, err, string(out))
	}
	return nil
}
