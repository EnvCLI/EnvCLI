package docker

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	isatty "github.com/mattn/go-isatty"
)

// Mounting volumes
func volumeMount(shellCommand *bytes.Buffer, mounts *[]ContainerMount) {
	for _, containerMount := range *mounts {
		if containerMount.MountType == "directory" {
			var mountSource = containerMount.Source
			var mountTarget = containerMount.Target
			// docker toolbox doesn't support direct mounts, so we have to use the shared folder feature
			if isDockerToolbox() && runtime.GOOS == "windows" {
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

// TODO: This function does more than one job.
//       Split this into another functions
func ContainerExec(image string, commandShell string, command string, mounts []ContainerMount, workingdir string, environment []string, publish []string) {
	var shellCommand bytes.Buffer

	// Shell (wrap the command within the container into a shell)
	command = sanitizeCommand(commandShell, command)
	// build docker command
	// - docker machine prefix
	if isDockerToolbox() {
		shellCommand.WriteString("docker-machine ssh envcli ")
	}
	// - docker
	shellCommand.WriteString("docker run --rm ")
	// - terminal
	_, ciVariableSet := os.LookupEnv("CI")
	if ciVariableSet {
		// env variable CI is set, we can't use --tty or --interactive here
	} else if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		shellCommand.WriteString("--tty --interactive ")
	}
	// - environment variables
	for _, envVariable := range environment {
		pair := strings.SplitN(envVariable, "=", 2)
		var envName = pair[0]
		var envValue = pair[1]

		shellCommand.WriteString(fmt.Sprintf("--env %s=%s ", envName, strconv.Quote(envValue)))
	}
	// - publish ports
	publishPorts(&shellCommand, &publish)
	// - set working directory
	shellCommand.WriteString(fmt.Sprintf("--workdir %s ", workingdir))
	// - volume mounts
	volumeMount(&shellCommand, &mounts)
	// - image
	shellCommand.WriteString(fmt.Sprintf("%s ", image))
	// - command to run inside of the container
	shellCommand.WriteString(fmt.Sprintf("%s", command))

	// execute command
	systemExec(shellCommand.String())
}
