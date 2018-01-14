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
 * CLI Command Passthru with input/output
 */
func systemExec(command string) {
	// Use Powershell on Windows
	if runtime.GOOS == "windows" {
		command = fmt.Sprintf("powershell %s", command)
	}

	log.Debugf("Running Command: %s", command)

	// Arguments
	args := strings.Fields(command)

	// Run Command
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to execute command: %s\n", err.Error())
		os.Exit(1)
	}
}

/**
 * Sets the loglevel according to the flag on each command run
 */
func setLoglevel(loglevel string) {
	if loglevel == "info" {
		log.SetLevel(log.InfoLevel)
	} else if loglevel == "debug" {
		log.SetLevel(log.DebugLevel)
	}
}
