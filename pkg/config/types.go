package config

// ConfigurationLoader contains all methods to load/save configuration files
type ConfigurationLoader struct {
}

// ConfigurationFile is the schema for configuration files, that hold multiple command specifications
type ConfigurationFile struct {
	Version string                  `yaml:"version" default:"v1"`
	Images  []RunConfigurationEntry `yaml:"images"`
}

// RunConfigurationEntry holds the configuration for a single command
type RunConfigurationEntry struct {
	// name of the container
	Name string `yaml:"name"`

	// description for the  container
	Description string `yaml:"description"`

	// the commands provided by the image
	Provides []string `yaml:"provides"`

	// container image
	Image string `yaml:"image"`

	// target directory to mount your project inside the container
	Directory string `default:"/project"`

	// overwrite the default entrypoint
	Entrypoint string `yaml:"entrypoint" default:"unset"`

	// wrap the executed command inside the container into a shell (ex. if you use globs)
	Shell string `yaml:"shell" default:"none"`

	// commands that should run in the container before the actual command is executed
	BeforeScript []string `yaml:"before_script"`

	// allows a container to access the container runtime on the host
	ContainerRuntimeAccess bool `yaml:"containerRuntimeAccess"`

	// add capabilities to the container
	CapAdd []string `yaml:"capAdd"`

	// Caching of container-directories
	Caching []CachingEntry `yaml:"cache"`

	// the command scope (internal use only) - global or project
	Scope string `yaml:"scope"`
}

type CachingEntry struct {

	/**
	 * Name of the caching entry
	 */
	Name string `yaml:"name" default:""`

	/**
	 * Directory inside the container that should be mounted on the host within the cache directory
	 */
	ContainerDirectory string `yaml:"directory" default:""`
}

type PropertyConfigurationFile struct {
	Properties map[string]string
}
