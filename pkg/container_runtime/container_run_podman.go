package container_runtime

import (
	"bytes"
	"fmt"
	"os"
	"strconv"

	isatty "github.com/mattn/go-isatty"
)

// GetPodmanCommand renders the command needed the run the container using podman
func (c *Container) GetPodmanCommand() string {
	var shellCommand bytes.Buffer

	// detect cygwin -> needs winpty on windows
	if isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		shellCommand.WriteString("winpty ")
	}

	// build command
	shellCommand.WriteString("podman run --rm ")
	// - terminal
	setTerminalParameters(&shellCommand)
	// - name
	if len(c.name) > 0 {
		shellCommand.WriteString(fmt.Sprintf("--name %s ", strconv.Quote(c.name)))
	}
	// - entrypoint
	if c.entrypoint != "unset" {
		shellCommand.WriteString(fmt.Sprintf("--entrypoint %s", strconv.Quote(c.entrypoint)))
	} else {
		shellCommand.WriteString("--entrypoint= ")
	}
	// - environment variables
	setEnvironmentVariables(&shellCommand, &c.environment)
	// - publish ports
	publishPorts(&shellCommand, &c.containerPorts)
	// - capabilities
	for _, cap := range c.capabilities {
		shellCommand.WriteString(fmt.Sprintf("--cap-add %s", strconv.Quote(cap)))
	}
	// - set working directory
	if len(c.workingDirectory) > 0 {
		shellCommand.WriteString(fmt.Sprintf("--workdir %s ", strconv.Quote(c.workingDirectory)))
	}
	// - volume mounts
	volumeMount(&shellCommand, &c.volumes)
	// - userArgs
	if len(c.userArgs) > 0 {
		shellCommand.WriteString(c.userArgs + " ")
	}
	// - image
	shellCommand.WriteString(fmt.Sprintf("%s ", c.image))
	// - command to run inside of the container
	shellCommand.WriteString(sanitizeCommand(c.commandShell, c.command))

	return shellCommand.String()
}
