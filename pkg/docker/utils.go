package docker

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus" // imports as package "log"
)

/**
 * Is CI Environment
 */
func isCIEnvironment() bool {
	_, ciVariableSet := os.LookupEnv("CI")
	if ciVariableSet {
		return true
	}

	return false
}

/**
 * Detect Docker native
 */
func isDockerNative() bool {
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
func isDockerToolbox() bool {
	path, err := exec.LookPath("docker-machine")
	if err != nil || strings.Contains(path, "Docker Toolbox") == false {
		return false
	}

	log.Debugf("Found Docker Toolbox at [%s].", path)
	return true
}

/**
 * Fix escaping character
 */
func sanitizeCommand(commandShell string, command string) string {
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

	return command
}

/**
 * CLI Command Passthru with input/output
 */
func systemExec(command string) error {
	log.Debugf("Running Command: %s", command)

	// Run Command
	if runtime.GOOS == "linux" {
		cmd := exec.Command("/usr/bin/env", "sh", "-c", command)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Fatalf("Failed to execute command: %s\n", err.Error())
			return err
		}
	} else if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", command)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Fatalf("Failed to execute command: %s\n", err.Error())
			return err
		}
	}

	return nil
}
