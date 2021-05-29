package container_runtime

import (
	"bytes"
	"fmt"
	"github.com/cidverse/cidverseutils/pkg/cihelper"
	"runtime"
	"strconv"
	"strings"
)

func setTerminalParameters(shellCommand *bytes.Buffer) {
	if cihelper.IsCIEnvironment() {
		// env variable CI is set, we can't use --tty or --interactive here
	} else if cihelper.IsInteractiveTerminal() {
		shellCommand.WriteString("-ti") // --tty --interactive
	}
}

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

			shellCommand.WriteString(fmt.Sprintf("-v %s ", strconv.Quote(mountSource+":"+mountTarget)))
		}
	}
}

func publishPorts(shellCommand *bytes.Buffer, publish *[]ContainerPort) {
	for _, publishVariable := range *publish {
		shellCommand.WriteString(fmt.Sprintf("-p %d:%d ", publishVariable.Source, publishVariable.Target))
	}
}

func setEnvironmentVariables(shellCommand *bytes.Buffer, environment *[]EnvironmentProperty) {
	for _, envVariable := range *environment {
		shellCommand.WriteString(fmt.Sprintf("-e %s=%s ", envVariable.Name, strconv.Quote(envVariable.Value)))
	}
}

// ToUnixPath turns a windows path into a unix path
func ToUnixPath(path string) string {
	driveLetters := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
	for _, element := range driveLetters {
		if strings.HasPrefix(path, element+":\\") {
			path = strings.Replace(path, element+":\\", "/"+element+"/", 1)
		}
	}

	// replace windows path separator with linux path separator
	path = strings.Replace(path, "\\", "/", -1)

	return path
}
