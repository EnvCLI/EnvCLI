package main

import (
	"fmt"
	"bytes"
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
func (docker Docker) containerExec(image string, tag string, commandShell string, command string, mountSource string, mountTarget string, workingdir string, environment []string) {
	var shellCommand bytes.Buffer

	// docker toolbox doesn't support direct mounts, so we have to use the shared folder feature
	if docker.isDockerToolbox() && runtime.GOOS == "windows" {
		driveLetters := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
		for _, element := range driveLetters {
			mountSource = strings.Replace(mountSource, element+":\\", "/"+element+"_DRIVE/", 1)
		}

		// replace windows path separator with linux path separator
		mountSource = strings.Replace(mountSource, "\\", "/", -1)
	}

	// Shell (wrap the command within the container into a shell)
	if commandShell == "powershell" {
		command = fmt.Sprintf("powershell %s", command)
	} else if commandShell == "sh" {
		command = strings.Replace(command, "\"", "\\\"", -1)
		command = fmt.Sprintf("/usr/bin/env sh -c \"%s\"", command)
	} else if commandShell == "bash" {
		command = strings.Replace(command, "\"", "\\\"", -1)
		command = fmt.Sprintf("/usr/bin/env bash -c \"%s\" -l", command)
	}

	// build docker command
	// - docker machine prefix
	if docker.isDockerToolbox() {
		shellCommand.WriteString("docker-machine ssh envcli ")
	}
	// - docker
	shellCommand.WriteString("docker run --rm --interactive --tty ")
	// - environment variables
	for _, envVariable := range environment {
		shellCommand.WriteString(fmt.Sprintf("--env %s ", envVariable))
	}
	// - set working directory
	shellCommand.WriteString(fmt.Sprintf("--workdir %s ", workingdir))
	// - volume mounts
	shellCommand.WriteString(fmt.Sprintf("--volume %s:%s ", mountSource, mountTarget))
	// - image
	shellCommand.WriteString(fmt.Sprintf("%s:%s ", image, tag))
	// - command to run inside of the container
	shellCommand.WriteString(fmt.Sprintf("%s", command))

	// execute command
  systemExec(shellCommand.String())
}
