package container_runtime

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	isatty "github.com/mattn/go-isatty"
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
	c.volumes = append(c.volumes, mount)
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

// GetRunCommand renders the command needed the run the container
func (c *Container) GetRunCommand() string {
	var shellCommand bytes.Buffer

	// build docker command
	// - docker
	shellCommand.WriteString("docker run --rm ")
	// - terminal
	if IsCIEnvironment() {
		// env variable CI is set, we can't use --tty or --interactive here
	} else if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		shellCommand.WriteString("-ti ") // tty + interactive
	}
	// - entrypoint
	if c.entrypoint != "original" {
		shellCommand.WriteString(fmt.Sprintf("--entrypoint %s", strconv.Quote(c.entrypoint)))
	}
	// - environment variables
	setEnvironmentVariables(&shellCommand, &c.environment)
	// - publish ports
	publishPorts(&shellCommand, &c.containerPorts)
	// - set working directory
	if len(c.workingDirectory) > 0 {
		shellCommand.WriteString(fmt.Sprintf("--workdir %s ", c.workingDirectory))
	}
	// - volume mounts
	volumeMount(&shellCommand, &c.volumes)
	// - image
	shellCommand.WriteString(fmt.Sprintf("%s ", c.image))
	// - command to run inside of the container
	shellCommand.WriteString(sanitizeCommand(c.commandShell, c.command))

	log.Debugf("Executed ShellCommand: %s", shellCommand.String())
	return shellCommand.String()
}

// StartContainer starts the Container
func (c *Container) StartContainer() {
	var shellCommand bytes.Buffer

	// - docker machine prefix
	if IsDockerToolbox() {
		shellCommand.WriteString("docker-machine ssh envcli ")
	}

	// - docker
	shellCommand.WriteString(c.GetRunCommand())

	// execute command
	systemExec(shellCommand.String())
}