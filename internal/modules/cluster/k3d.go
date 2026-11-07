package cluster

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// K3dManager manages local k3d clusters.
type K3dManager struct{}

// NewK3dManager creates a new K3dManager.
func NewK3dManager() *K3dManager {
	return &K3dManager{}
}

// CheckAvailable checks if k3d and docker are installed and working.
func (m *K3dManager) CheckAvailable() (bool, string) {
	if _, err := exec.LookPath("docker"); err != nil {
		return false, "docker is not installed or not in PATH"
	}
	if _, err := exec.LookPath("k3d"); err != nil {
		return false, "k3d is not installed or not in PATH"
	}
	return true, ""
}

// ClusterInfo holds basic info about a detected cluster.
type ClusterInfo struct {
	Name    string
	Servers string
	Agents  string
	Status  string
}

// ListClusters returns a list of k3d clusters.
func (m *K3dManager) ListClusters(ctx context.Context) ([]ClusterInfo, error) {
	available, reason := m.CheckAvailable()
	if !available {
		return nil, fmt.Errorf("k3d unavailable: %s", reason)
	}

	cmd := exec.CommandContext(ctx, "k3d", "cluster", "list", "--no-headers")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed listing k3d clusters: %w", err)
	}

	var list []ClusterInfo
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 1 {
			info := ClusterInfo{Name: fields[0]}
			if len(fields) >= 3 {
				info.Servers = fields[1]
				info.Agents = fields[2]
			}
			list = append(list, info)
		}
	}
	return list, nil
}

// CreateCluster creates a new k3d cluster.
func (m *K3dManager) CreateCluster(ctx context.Context, name string, ephemeral bool) error {
	available, reason := m.CheckAvailable()
	if !available {
		return fmt.Errorf("cannot create cluster: %s", reason)
	}

	args := []string{"cluster", "create", name, "--servers", "1", "--wait"}
	cmd := exec.CommandContext(ctx, "k3d", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("k3d cluster create failed: %s (%w)", string(out), err)
	}
	return nil
}

// DeleteCluster removes a k3d cluster.
func (m *K3dManager) DeleteCluster(ctx context.Context, name string) error {
	available, reason := m.CheckAvailable()
	if !available {
		return fmt.Errorf("cannot delete cluster: %s", reason)
	}

	cmd := exec.CommandContext(ctx, "k3d", "cluster", "delete", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("k3d cluster delete failed: %s (%w)", string(out), err)
	}
	return nil
}

// RunEphemeralTest runs an ephemeral test workflow on a temporary k3d cluster.
func (m *K3dManager) RunEphemeralTest(ctx context.Context, testFn func(contextName string) error) error {
	available, reason := m.CheckAvailable()
	if !available {
		return fmt.Errorf("ephemeral testing requires k3d and docker: %s", reason)
	}

	clusterName := fmt.Sprintf("moreops-test-%d", time.Now().Unix())
	contextName := "k3d-" + clusterName

	fmt.Printf("🚀 Spinning up ephemeral k3d cluster '%s'...\n", clusterName)
	if err := m.CreateCluster(ctx, clusterName, true); err != nil {
		return err
	}

	// Always cleanup the ephemeral cluster
	defer func() {
		fmt.Printf("🧹 Tearing down ephemeral k3d cluster '%s'...\n", clusterName)
		_ = m.DeleteCluster(context.Background(), clusterName)
	}()

	fmt.Printf("✓ Ephemeral cluster ready (context: %s). Running tests...\n", contextName)
	return testFn(contextName)
}
