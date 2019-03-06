package main

import (
	log "github.com/sirupsen/logrus" // imports as package "log"
	"os"
	"strings"
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
 * Create folders if they don't exist
 */
func createDirectory(dir string) {
	os.MkdirAll(dir, os.ModePerm)
}

/**
 * Get the working directory
 */
func getWorkingDirectory() string {
	workingDir, err := os.Getwd()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Couldn't detect working directory!")
	}

	return workingDir
}

/**
 * Get the relative path in relation to the rootDirectory
 */
func getPathRelativeToDirectory(currentDirectory string, rootDirectory string) string {
	relativePath := strings.Replace(currentDirectory, rootDirectory, "", 1)
	relativePath = strings.Replace(relativePath, "\\", "/", -1)
	relativePath = strings.Trim(relativePath, "/")

	return relativePath
}

/**
 * Gets the value or the specified default value if not found or empty
 */
func getOrDefault(entity map[string]string, key string, defaultValue string) (val string) {
	value, found := entity[key]

	if found {
		return value
	}

	return defaultValue
}

/**
 * Detect CI
 */
func detectCIEnvironment() (val bool) {
	value, found := os.LookupEnv("CI")
	if found && value == "true" {
		return true
	}

	return false
}
