package docker

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
