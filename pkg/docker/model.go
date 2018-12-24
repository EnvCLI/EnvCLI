package docker

/**
 * All functions to interact with docker
 */
type Docker struct {
}

type ContainerMount struct {

	/**
	 * Mount Type
	 */
	MountType string

	/**
	 * Source Directory (Host) / Source Volume
	 */
	Source string

	/**
	 * Target Directory (Container)
	 */
	Target string
}
