package container_runtime

import (
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus" // imports as package "log"
)

/**
 * Is CI Environment
 */
func IsCIEnvironment() bool {
	// usually set by ci
	_, ciVariableSet := os.LookupEnv("CI")
	if ciVariableSet {
		return true
	}

	// set by normalize ci
	_, nciVariableSet := os.LookupEnv("NCI")
	if nciVariableSet {
		return true
	}

	return false
}

/**
 * Detect Podman
 */
func IsPodman() bool {
	path, err := exec.LookPath("podman")
	if err != nil {
		return false
	}

	log.Tracef("Found Podman at [%s].", path)
	return true
}

/**
 * Detect Docker native
 */
func IsDockerNative() bool {
	path, err := exec.LookPath("docker")
	if err != nil {
		return false
	}

	log.Tracef("Found Docker native at [%s].", path)
	return true
}

// IsDockerToolbox returns true, if docker toolbox is used
func IsDockerToolbox() bool {
	path, err := exec.LookPath("docker-machine")
	if err != nil || strings.Contains(path, "Docker Toolbox") == false {
		return false
	}

	log.Tracef("Found Docker Toolbox at [%s].", path)
	return true
}

// IsMinGW returns true, if the binary is called from a Minimalist GNU for Windows environment (cygwin / git bash)
func IsMinGW() bool {
	value, _ := os.LookupEnv("MSYSTEM")
	if value == "MINGW64" {
		return true
	}

	return false
}

/**
 * Fix escaping character
 */
func sanitizeCommand(commandShell string, command string) string {
	// Shell (wrap the command within the container into a shell)
	if commandShell == "powershell" {
		// would be used for windows containers, never tested though
		command = fmt.Sprintf("powershell %s", command)
	} else if commandShell == "sh" {
		if runtime.GOOS == "windows" {
			command = fmt.Sprintf("\"/usr/bin/env\" \"sh\" \"-c\" \"%s\"", strings.Replace(command, "\"", "`\"", -1))
		} else {
			command = fmt.Sprintf("\"/usr/bin/env\" \"sh\" \"-c\" \"%s\"", strings.Replace(command, "\"", "\\\"", -1))
		}
	} else if commandShell == "bash" {
		if runtime.GOOS == "windows" {
			command = fmt.Sprintf("\"/usr/bin/env\" \"bash\" \"-l\" \"-c\" \"%s\"", strings.Replace(command, "\"", "`\"", -1))
		} else {
			command = fmt.Sprintf("\"/usr/bin/env\" \"bash\" \"-l\" \"-c\" \"%s\"", strings.Replace(command, "\"", "\\\"", -1))
		}
	}

	return command
}

/**
 * CLI Command Passthru with input/output
 */
func systemExec(command string) error {
	// Run Command
	if runtime.GOOS == "linux" {
		cmd := exec.Command("/usr/bin/env", "sh", "-c", command)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		log.Debugf("Running Command: /usr/bin/env sh -c %s", command)
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
		log.Debugf("Running Command: powershell %s", command)
		if err != nil {
			log.Fatalf("Failed to execute command: %s\n", err.Error())
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
