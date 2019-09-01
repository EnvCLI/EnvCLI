package container_runtime

// ContainerRuntime
type ContainerRuntime struct{}

// NewContainer will get a new container struct to work with
func (cr *ContainerRuntime) NewContainer() *Container {
	container := &Container{}
	container.SetEntrypoint("unset")
	return container
}
