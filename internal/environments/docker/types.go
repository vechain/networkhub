package docker

// ExposedPort represents a port mapping between host and container
type ExposedPort struct {
	HostPort      string
	ContainerPort string
}
