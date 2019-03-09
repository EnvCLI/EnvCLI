package main

import (
	log "github.com/sirupsen/logrus" // imports as package "log"
	"os"
)

/**
 * Sets the loglevel according to the flag on each command run
 */
func setLoglevel(loglevel string) {
	if loglevel == "panic" {
		log.SetLevel(log.PanicLevel)
	} else if loglevel == "fatal" {
		log.SetLevel(log.FatalLevel)
	} else if loglevel == "warn" {
		log.SetLevel(log.WarnLevel)
	} else if loglevel == "info" {
		log.SetLevel(log.InfoLevel)
	} else if loglevel == "debug" {
		log.SetLevel(log.DebugLevel)
	} else if loglevel == "trace" {
		log.SetLevel(log.TraceLevel)
	}
}

/**
 * Detect CI
 */
func DetectCIEnvironment() (val bool) {
	value, found := os.LookupEnv("CI")
	if found && value == "true" {
		return true
	}

	return false
}
