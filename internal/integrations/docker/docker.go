package docker

// DockerIntegration handles Docker CLI invocations and daemon connections.
type DockerIntegration struct {
	SocketPath string
}

func NewDockerIntegration() *DockerIntegration {
	return &DockerIntegration{
		SocketPath: "/var/run/docker.sock",
	}
}
