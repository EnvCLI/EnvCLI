package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"runtime"
	log "github.com/sirupsen/logrus" // imports as package "log"
)

/**
 * All functions to interact with docker
 */
type Docker struct {
}

type DockerMount struct {
	SourceDirectory string
	TargetDirectory string
}

/**
 * Detect Docker native
 */
func (docker Docker) isDockerNative() bool {
	path, err := exec.LookPath("docker")
	if err != nil {
			return false
	} else {
		log.Debugf("Found Docker native at [%s].", path)
		return true
	}
}

/**
 * Detect Docker Toolbox
 */
func (docker Docker) isDockerToolbox() bool {
	path, err := exec.LookPath("docker-machine")
	if err != nil || strings.Contains(path, "Docker Toolbox") == false {
			return false
	} else {
		log.Debugf("Found Docker Toolbox at [%s].", path)
		return true
	}
}

/**
 * Run a Command in Docker
 */
func (docker Docker) containerExec(image string, tag string, command string, mountSource string, mountTarget string, workingdir string) {
	// docker toolbox doesn't support direct mounts, so we have to use the shared folder feature
	if docker.isDockerToolbox() && runtime.GOOS == "windows" {
		log.Debugf("Replacement for [%s].", mountSource)
		driveLetters := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
		for _, element := range driveLetters {
			mountSource = strings.Replace(mountSource, element + ":\\", "/" + element + "_DRIVE/", 1)
		}

		// replace windows path seperator with linux path seperator
		mountSource = strings.Replace(mountSource, "\\", "/", -1)
	}

	var commandPrefix string = ""
	if docker.isDockerToolbox() {
		commandPrefix = "docker-machine ssh envcli "
	}

	var dockerCommand string = fmt.Sprintf("%sdocker run --rm --interactive --tty --workdir %s --volume \"%s:%s\" %s:%s %s", commandPrefix, workingdir, mountSource, mountTarget, image, tag, command)
	execCommandWithResponse(dockerCommand)
}

/**
 * CLI Command Passthru with input/output
 */
func execCommandWithResponse(command string) {
	var commandPrefix string
	if runtime.GOOS == "windows" {
		commandPrefix = "powershell"
	} else {
		commandPrefix = ""
	}

	log.Debugf("Running Command: %s", command)

	cmd := exec.Command(commandPrefix, command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
  if err != nil {
			log.Fatalf("Failed to execute command: %s\n", err.Error())
      os.Exit(1)
  }
}
