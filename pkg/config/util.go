package config

import (
	"os"

	log "github.com/sirupsen/logrus"
)

/**
 * Get the working directory
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
