package infrastructure

import "moreops/internal/modules/cluster"

type Manager struct {
	k3d *cluster.K3dManager
}

func NewManager() *Manager {
	return &Manager{
		k3d: cluster.NewK3dManager(),
	}
}
