package docker

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

// Mounting volumes
func volumeMount(shellCommand *bytes.Buffer, mounts *[]ContainerMount) {
	for _, containerMount := range *mounts {
		if containerMount.MountType == "directory" {
			var mountSource = containerMount.Source
			var mountTarget = containerMount.Target
			// docker toolbox doesn't support direct mounts, so we have to use the shared folder feature
			if IsDockerToolbox() && runtime.GOOS == "windows" {
				driveLetters := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
				for _, element := range driveLetters {
					mountSource = strings.Replace(mountSource, element+":\\", "/"+element+"_DRIVE/", 1)
				}

				// replace windows path separator with linux path separator
				mountSource = strings.Replace(mountSource, "\\", "/", -1)
			}

			shellCommand.WriteString(fmt.Sprintf("--volume \"%s:%s\" ", mountSource, mountTarget))
		}
	}

}

func publishPorts(shellCommand *bytes.Buffer, publish *[]string) {
	for _, publishVariable := range *publish {
		shellCommand.WriteString(fmt.Sprintf("--publish %s ", publishVariable))
	}

}

func setEnvironmentVariables(shellCommand *bytes.Buffer, environment *[]string) {
	for _, envVariable := range *environment {
		pair := strings.SplitN(envVariable, "=", 2)
		var envName = pair[0]
		var envValue = pair[1]

		shellCommand.WriteString(fmt.Sprintf("--env %s=%s ", envName, strconv.Quote(envValue)))
	}
}

func setTerminalParameters(shellCommand *bytes.Buffer) {
	if IsCIEnvironment() {
		// env variable CI is set, we can't use --tty or --interactive here
	} else if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		shellCommand.WriteString("--tty --interactive ")
	}
}

func setEntrypoint(shellCommand *bytes.Buffer, entrypoint *string) {
	if *entrypoint != "original" {
		shellCommand.WriteString(fmt.Sprintf("--entrypoint %s", strconv.Quote(*entrypoint)))
	}
}

/**
 * ContainerExec runs a container and attaches the result to the current session
 */
func ContainerExec(image string, entrypoint string, commandShell string, command string, mounts []ContainerMount, workingdir string, environment []string, publish []string) {
	var shellCommand bytes.Buffer

	// build docker command
	// - docker
	shellCommand.WriteString("docker run --rm ")
	// - terminal
	setTerminalParameters(&shellCommand)
	// - entrypoint
	setEntrypoint(&shellCommand, &entrypoint)
	// - environment variables
	setEnvironmentVariables(&shellCommand, &environment)
	// - publish ports
	publishPorts(&shellCommand, &publish)
	// - set working directory
	shellCommand.WriteString(fmt.Sprintf("--workdir %s ", workingdir))
	// - volume mounts
	volumeMount(&shellCommand, &mounts)
	// - image
	shellCommand.WriteString(fmt.Sprintf("%s ", image))
	// - command to run inside of the container
	command = sanitizeCommand(commandShell, command)
	shellCommand.WriteString(fmt.Sprintf("%s", command))

	log.Debugf("Executed ShellCommand: %s", shellCommand.String())

	// execute command
	RunDockerCommand(shellCommand.String())
}

/**
 * RunDockerCommand executes a docker command and wrap's it for docker machine if needed
 */
func RunDockerCommand(command string) {
	var shellCommand bytes.Buffer

	// - docker machine prefix
	if IsDockerToolbox() {
		shellCommand.WriteString("docker-machine ssh envcli ")
	}
	// - docker
	shellCommand.WriteString(command)

	// execute command
	systemExec(shellCommand.String())
}
