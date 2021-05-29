package util

import (
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strings"
)

// CreateDirectory creates a new folder if not present, ignores errors
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
		log.Fatal().Err(err).Msg("Couldn't detect execution directory!")
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
		log.Fatal().Err(err).Msg("Couldn't detect working directory!")
	}

	return workingDir
}
