package main

import (
	"os"
	log "github.com/sirupsen/logrus" // imports as package "log"
	"github.com/jinzhu/configor" // imports as package "configor"
)

// TODO: Check for configuration in parent directories

/**
 * The Project Configuration and it's properties
 */
type ConfigurationLoader struct {
}

/**
 * The Project Configuration
 */
type ProjectConfigrationFile struct {
	Commands []struct {
		Name  string
		Description string
		Image string
		Tag string
		ProjectDirectory string `default:"/project"`
	}
}

/**
 * Load the .devcli.yml Configuration
 */
func (configurationLoader ConfigurationLoader) load(configFile string) ProjectConfigrationFile {
	var cfg ProjectConfigrationFile;

	log.Debug("Loading project configuration file " + configFile)
	configor.New(&configor.Config{Debug: false}).Load(&cfg, configFile)

 	return cfg
}

/**
 * Get the working directory
 */
func (configurationLoader ConfigurationLoader) getWorkingDirectory() string {
	workingDir, err := os.Getwd()
	if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("Couldn't detect working directory!")
	}

 	return workingDir
}
