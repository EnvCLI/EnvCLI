package main

import (
	log "github.com/sirupsen/logrus" // imports as package "log"
	"io"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
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
 * Checks if a object is part of a array
 */
func inArray(val interface{}, array interface{}) (exists bool, index int) {
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

/**
 * DownloadFile will download a url to a local file.
 */
func DownloadFile(filepath string, url string) error {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
