package container_runtime

import (
	"bytes"
	"errors"
	"github.com/cidverse/cidverseutils/pkg/cihelper"
	"github.com/cidverse/cidverseutils/pkg/collection"
	"github.com/rs/zerolog/log"
	"os"
	"runtime"
	"strconv"
	"strings"
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
	mount.Source = ToUnixPath(mount.Source)

	// modify mount source on MinGW environments
	if cihelper.IsMinGW() {
		// git bash / cygwin needs the host path escaped with a leading / -> //c so that it works correctly
		mount.Source = "/"+mount.Source
	}

	c.volumes = append(c.volumes, mount)
}

// AddCacheMount adds a cache mount to the container
func (c *Container) AddCacheMount(name string, sourcePath string, targetPath string) {
	c.AddVolume(ContainerMount{MountType: "directory", Source: ToUnixPath(sourcePath), Target: targetPath})
	c.AddEnvironmentVariable("cache_"+name+"_source", ToUnixPath(sourcePath))
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

	// MinGW environments
	if cihelper.IsMinGW() {
		// git bash / cygwin needs the host path escaped with a leading / -> //c so that it works correctly
		c.workingDirectory = "/"+c.workingDirectory
	}
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
		var systemVars = []string{
			"",
			// unix
			"_",
			"PWD",
			"OLDPWD",
			"PATH",
			"HOME",
			"HOSTNAME",
			"TERM",
			"SHLVL",
			// windows
			"PROGRAMDATA",
			"PROGRAMFILES",
			"PROGRAMFILES(x86)", 
			"PROGRAMW6432",
			"COMMONPROGRAMFILES",
			"COMMONPROGRAMFILES(x86)",
			"COMMONPROGRAMW6432",
			// proxy
			"HTTP_PROXY",
			"HTTPS_PROXY",
		}
		isExcluded, _ := collection.InArray(strings.ToUpper(envName), systemVars)
		// recent issue of 2009 about git bash / mingw setting invalid unix variables with `var(86)=...`
		isInvalidName := strings.Contains(envName, "(") || strings.Contains(envName, ")")
		if !isExcluded && !isInvalidName {
			log.Debug().Msg("Added environment variable "+envName+" ["+envValue+"] from host!")
			c.AddEnvironmentVariable(envName, envValue)
		} else if !isExcluded {
			log.Debug().Msg("Excluded env variable "+envName+" ["+envValue+"]  from host because of a invalid variable name.")
		} else {
			log.Debug().Msg("Excluded env variable "+envName+" ["+envValue+"]  from host based on the filter rule.")
		}
	}
}

// SetUserArgs allows the user to pass custom arguments to the container run command, for special cases in ci envs with service links / or similar
func (c *Container) SetUserArgs(newArgs string) {
	c.userArgs = newArgs
}

// DetectRuntime returns the first available container runtime
func (c *Container) DetectRuntime() string {
	// autodetect container runtime
	if IsPodman() {
		return "podman"
	} else if IsDockerNative() || IsDockerToolbox() {
		return "docker"
	}

	return "unknown"
}

// GetPullCommand gets the command to pull the required image
func (c *Container) GetPullCommand(runtime string) (string, error) {
	// autodetect container runtime
	if runtime == "podman" {
		return "podman pull "+c.image, nil
	} else if runtime == "docker" {
		return "docker pull "+c.image, nil
	} else {
		return "", errors.New("No supported container runtime found (podman, docker, docker toolbox)! ["+runtime+"]")
	}
}

// GetRunCommand gets the run command for the specified container runtime
func (c *Container) GetRunCommand(runtime string) string {
	var shellCommand bytes.Buffer

	// autodetect container runtime
	if runtime == "podman" {
		shellCommand.WriteString(c.GetPodmanCommand())
	} else if runtime == "docker" {
		shellCommand.WriteString(c.GetDockerCommand())
	} else {
		log.Fatal().Str("runtime", runtime).Msg("Container Runtime is not supported!")
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
	shellCommand.WriteString(c.GetRunCommand(c.DetectRuntime()))

	// execute command
	systemExec(shellCommand.String())
}

// PullImage pulls the image for the container
func (c *Container) PullImage() {
	pullCmd, pullCmdErr := c.GetPullCommand(c.DetectRuntime())
	if pullCmdErr == nil {
		systemExec(pullCmd)
	} else {
		log.Error().Err(pullCmdErr).Msg("Can't pull image")
		os.Exit(1)
	}
}

