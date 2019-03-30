package container_runtime

import (
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"

	sentry "github.com/EnvCLI/EnvCLI/pkg/sentry"
	log "github.com/sirupsen/logrus" // imports as package "log"
)

/**
 * Is CI Environment
 */
func IsCIEnvironment() bool {
	_, ciVariableSet := os.LookupEnv("CI")
	if ciVariableSet {
		return true
	}

	return false
}

/**
 * Detect Docker native
 */
func IsDockerNative() bool {
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
func IsDockerToolbox() bool {
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
		command = fmt.Sprintf("/usr/bin/env sh -c \"%s\"", command)
	} else if commandShell == "bash" {
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
			sentry.HandleError(err)
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
			sentry.HandleError(err)
			return err
		}
	}

	return nil
}

/**
 * Checks if a object is part of a array
 */
func InArray(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}

	return
}