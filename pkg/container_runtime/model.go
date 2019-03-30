package container_runtime

// ContainerMount holds container volume mounts
type ContainerMount struct {
	MountType string
	Source    string
	Target    string
}

// EnvironmentProperty holds environment variables
type EnvironmentProperty struct {
	Name  string
	Value string
}

// ContainerPort holds container ports
type ContainerPort struct {
	Source int
	Target int
}
