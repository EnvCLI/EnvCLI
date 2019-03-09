package util

import (
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

/**
 * Create folders if they don't exist
 */
func CreateDirectory(dir string) {
	os.MkdirAll(dir, os.ModePerm)
}

/**
 * Get the relative path in relation to the rootDirectory
 */
func GetPathRelativeToDirectory(currentDirectory string, rootDirectory string) string {
	relativePath := strings.Replace(currentDirectory, rootDirectory, "", 1)
	relativePath = strings.Replace(relativePath, "\\", "/", -1)
	relativePath = strings.Trim(relativePath, "/")

	return relativePath
}

/**
 * Get the execution directory
 */
func GetExecutionDirectory() string {
	ex, err := os.Executable()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Couldn't detect execution directory!")
		return ""
	}

	return filepath.Dir(ex)
}

/**
 * GetWorkingDirectory gets the current working directory
 */
func GetWorkingDirectory() string {
	workingDir, err := os.Getwd()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Couldn't detect working directory!")
	}

	return workingDir
}
