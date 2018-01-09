package main

import (
	"github.com/jinzhu/configor"     // imports as package "configor"
	log "github.com/sirupsen/logrus" // imports as package "log"
	yaml "gopkg.in/yaml.v2"          // imports as package "yaml"
	"os"
	"strings"
	"path/filepath"
	"io/ioutil"
)

/**
 * The Project Configuration and it's properties
 */
type ConfigurationLoader struct {
}

/**
 * The Project Configuration
 */
type ProjectConfigrationFile struct {
	DockerMachineVM string `default:"envcli"`
	Commands        []struct {
		Name        string
		Description string
		Provides    []string
		Image       string
		Tag         string
		Directory   string `default:"/project"`
		Shell       string `default:"none"`
	}
}

/**
 * The EnvCLI Configuration
 */
type CLIConfigrationFile struct {
	HttpProxy string `default:""`
	HttpsProxy string `default:""`
}

/**
 * Load the project config
 */
func (configurationLoader ConfigurationLoader) loadProjectConfig(configFile string) ProjectConfigrationFile {
	var cfg ProjectConfigrationFile

	log.Debug("Loading project configuration file " + configFile)
	configor.New(&configor.Config{Debug: false}).Load(&cfg, configFile)

	return cfg
}

/**
 * Load the global config
 */
func (configurationLoader ConfigurationLoader) loadGlobalConfig(configFile string) CLIConfigrationFile {
	var cfg CLIConfigrationFile

	log.Debug("Loading global configuration file " + configFile)
	configor.New(&configor.Config{Debug: false}).Load(&cfg, configFile)

	return cfg
}

/**
 * Save the global config
 */
func (configurationLoader ConfigurationLoader) saveGlobalConfig(configFile string, cfg CLIConfigrationFile) error {
	log.Debug("Saving global configuration file " + configFile)

	fileContent, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(configFile, fileContent, 0600)
}

/**
 * Get the execution directory
 */
func (configurationLoader ConfigurationLoader) getExecutionDirectory() string {
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

/**
 * Get the project root directory by searching for the envcli config
 */
func (configurationLoader ConfigurationLoader) getProjectDirectory() string {
	currentDirectory := configurationLoader.getWorkingDirectory()
	var projectDirectory string = ""

	directoryParts := strings.Split(currentDirectory, string(os.PathSeparator))

	for projectDirectory == "" {
		if _, err := os.Stat(filepath.Join(currentDirectory, "/.envcli.yml")); err == nil {
			return currentDirectory
		} else {
			if directoryParts[0]+"\\" == currentDirectory {
				return ""
			}

			currentDirectory = filepath.Dir(currentDirectory)
		}
	}

	return ""
}

/**
 * Get the relative path of the project directory to the current working directory
 */
func (configurationLoader ConfigurationLoader) getRelativePathToWorkingDirectory() string {
	currentDirectory := configurationLoader.getWorkingDirectory()
	projectDirectory := configurationLoader.getProjectDirectory()

	relativePath := strings.Replace(currentDirectory, projectDirectory, "", 1)
	relativePath = strings.Replace(relativePath, "\\", "/", -1)
	relativePath = strings.Trim(relativePath, "/")

	return relativePath
}
