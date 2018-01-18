package main

import (
	"os"
	"os/exec"
	"runtime"
	log "github.com/sirupsen/logrus" // imports as package "log"
)

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
