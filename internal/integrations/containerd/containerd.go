package containerd

// ContainerdIntegration handles containerd runtime configurations.
type ContainerdIntegration struct {
	SocketPath string
}

func NewContainerdIntegration() *ContainerdIntegration {
	return &ContainerdIntegration{
		SocketPath: "/run/containerd/containerd.sock",
	}
}
