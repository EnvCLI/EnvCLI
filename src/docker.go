package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	isatty "github.com/mattn/go-isatty"
	log "github.com/sirupsen/logrus" // imports as package "log"
)

/**
 * All functions to interact with docker
 */
type Docker struct {
}

type ContainerMount struct {

	/**
	 * Mount Type
	 */
	mountType string

	/**
	 * Source Directory (Host) / Source Volume
	 */
	source string

	/**
	 * Target Directory (Container)
	 */
	target string
}

/**
 * Detect Docker native
 */
func (docker Docker) isDockerNative() bool {
	path, err := exec.LookPath("docker")
	if err != nil {
		return false
	}

	log.Debugf("Found Docker native at [%s].", path)
	return true
}

/**
 * Detect Docker Toolbox
 */
func (docker Docker) isDockerToolbox() bool {
	path, err := exec.LookPath("docker-machine")
	if err != nil || strings.Contains(path, "Docker Toolbox") == false {
		return false
	}

	log.Debugf("Found Docker Toolbox at [%s].", path)
	return true
}

/**
 * Run a Command in Docker
 */
func (docker Docker) containerExec(image string, commandShell string, command string, mounts []ContainerMount, workingdir string, environment []string, publish []string) {
	var shellCommand bytes.Buffer

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
	for _, publishVariable := range publish {
		shellCommand.WriteString(fmt.Sprintf("--publish %s ", publishVariable))
	}
	// - set working directory
	shellCommand.WriteString(fmt.Sprintf("--workdir %s ", workingdir))
	// - volume mounts
	for _, containerMount := range mounts {
		if containerMount.mountType == "directory" {
			var mountSource = containerMount.source
			var mountTarget = containerMount.target
			// docker toolbox doesn't support direct mounts, so we have to use the shared folder feature
			if docker.isDockerToolbox() && runtime.GOOS == "windows" {
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

	// - image
	shellCommand.WriteString(fmt.Sprintf("%s ", image))
	// - command to run inside of the container
	shellCommand.WriteString(fmt.Sprintf("%s", command))

	// execute command
	systemExec(shellCommand.String())
}
