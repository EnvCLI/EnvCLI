package container_runtime

import (
	"bytes"
	"os"
	"runtime"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Container provides all methods to interact with the container runtime
type Container struct {
	name             string
	isRunning        bool
	image            string
	entrypoint       string
	commandShell     string
	command          string
	workingDirectory string
	volumes          []ContainerMount
	environment      []EnvironmentProperty
	containerPorts   []ContainerPort
	capabilities     []string
	userArgs         string
}

// SetName sets a new name for the container
func (c *Container) SetName(newName string) {
	c.name = newName
}

// GetName gets the container name
func (c *Container) GetName() string {
	return c.name
}

// SetImage sets the container image
func (c *Container) SetImage(newImage string) {
	c.image = newImage
}

// AddVolume mounts a directory into a container
func (c *Container) AddVolume(mount ContainerMount) {
	mount.Source = toUnixPath(mount.Source)
	c.volumes = append(c.volumes, mount)
}

// AddCacheMount adds a cache mount to the container
func (c *Container) AddCacheMount(name string, sourcePath string, targetPath string) {
	c.AddVolume(ContainerMount{MountType: "directory", Source: toUnixPath(sourcePath), Target: targetPath})
	c.AddEnvironmentVariable("cache_"+name+"_source", toUnixPath(sourcePath))
	c.AddEnvironmentVariable("cache_"+name+"_target", targetPath)
}

// AllowContainerRuntimeAcccess allows the container to access the container runtime
func (c *Container) AllowContainerRuntimeAcccess() {
	socketPath := "/var/run/docker.sock"
	if runtime.GOOS == "windows" {
		if IsDockerNative() {
			// docker desktop
			socketPath = "//var/run/docker.sock"
		}
	}

	c.AddVolume(ContainerMount{MountType: "directory", Source: socketPath, Target: "/var/run/docker.sock"})
}

// SetEntrypoint overwrites the default entrypoint
func (c *Container) SetEntrypoint(newEntrypoint string) {
	c.entrypoint = newEntrypoint
}

// SetCommandShell sets the command shell
func (c *Container) SetCommandShell(newCommandShell string) {
	c.commandShell = newCommandShell
}

// SetCommand sets the container command
func (c *Container) SetCommand(newCommand string) {
	c.command = newCommand
}

// SetWorkingDirectory sets the working directory
func (c *Container) SetWorkingDirectory(newWorkingDirectory string) {
	c.workingDirectory = newWorkingDirectory
}

// AddContainerPort publishes a port
func (c *Container) AddContainerPort(port ContainerPort) {
	c.containerPorts = append(c.containerPorts, port)
}

// AddCapability adds a capability to the container
func (c *Container) AddCapability(capability string) {
	c.capabilities = append(c.capabilities, capability)
}

// AddContainerPorts adds multiple published ports
func (c *Container) AddContainerPorts(ports []string) {
	for _, p := range ports {
		pair := strings.SplitN(p, ":", 2)
		sourcePort, _ := strconv.Atoi(pair[0])
		targetPort, _ := strconv.Atoi(pair[1])

		c.AddContainerPort(ContainerPort{Source: sourcePort, Target: targetPort})
	}
}

// AddEnvironmentVariable adds a environment variable
func (c *Container) AddEnvironmentVariable(name string, value string) {
	c.environment = append(c.environment, EnvironmentProperty{Name: name, Value: value})
}

// AddEnvironmentVariables adds multiple environment variables
func (c *Container) AddEnvironmentVariables(variables []string) {
	for _, e := range variables {
		pair := strings.SplitN(e, "=", 2)
		var envName = pair[0]
		var envValue = pair[1]

		c.AddEnvironmentVariable(envName, envValue)
	}
}

// AddAllEnvironmentVariables adds all environment variables, but filters a few irrelevant ones (like PATH, HOME, etc.)
func (c *Container) AddAllEnvironmentVariables() {
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		var envName = pair[0]
		var envValue = pair[1]

		// filter vars
		var systemVars = []string{"_", "PWD", "OLDPWD", "PATH", "HOME", "HOSTNAME", "TERM", "SHLVL", "HTTP_PROXY", "HTTPS_PROXY"}
		isExluded, _ := InArray(strings.ToUpper(envName), systemVars)
		if !isExluded {
			log.Debugf("Added environment variable %s [%s] from host!", envName, envValue)
			c.AddEnvironmentVariable(envName, envValue)
		} else {
			log.Debugf("Excluded env variable %s [%s] from host based on the filter rule.", envName, envValue)
		}
	}
}

// SetUserArgs allows the user to pass custom arguments to the container run command, for special cases in ci envs with service links / or similar
func (c *Container) SetUserArgs(newArgs string) {
	c.userArgs = newArgs
}

// GetRunCommand renders the command needed the run the container
func (c *Container) GetRunCommand() string {
	var shellCommand bytes.Buffer

	// autodetect container runtime
	if IsPodman() {
		shellCommand.WriteString(c.GetPodmanCommand())
	} else if IsDockerNative() || IsDockerToolbox() {
		shellCommand.WriteString(c.GetDockerCommand())
	} else {
		log.Fatalf("No supported container runtime found (podman, docker, docker toolbox)!")
	}

	return shellCommand.String()
}

// StartContainer starts the Container
func (c *Container) StartContainer() {
	var shellCommand bytes.Buffer

	// - workaround for docker toolbox (will be deprecated and removed from envcli when WSL 2 is released)
	if IsDockerToolbox() {
		shellCommand.WriteString("docker-machine ssh envcli ")
	}

	// - command
	shellCommand.WriteString(c.GetRunCommand())

	// execute command
	log.Debugf("Executed ShellCommand: %s", shellCommand.String())
	systemExec(shellCommand.String())
}
