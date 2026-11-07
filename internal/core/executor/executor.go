package executor

import (
	"bytes"
	"context"
	"os/exec"
)

type ShellExecutor struct{}

func New() *ShellExecutor {
	return &ShellExecutor{}
}

func (e *ShellExecutor) Run(ctx context.Context, cmdStr string) (string, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}
