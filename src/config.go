package main

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/jinzhu/configor"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

/**
 * ConfigurationLoader contains all methods to load/save configuration files
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
type PropertyConfigurationFile struct {
	HTTPProxy  string `default:""`
	HTTPSProxy string `default:""`
}

/**
 * Load the project config
 */
func (configurationLoader ConfigurationLoader) loadProjectConfig(configFile string) (ProjectConfigrationFile, error) {
	var cfg ProjectConfigrationFile

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return cfg, errors.New("project configuration file not found")
	}

	log.Debug("Loading project configuration file " + configFile)
	configor.New(&configor.Config{Debug: false}).Load(&cfg, configFile)

	return cfg, nil
}

/**
 * Load the global config
 */
func (configurationLoader ConfigurationLoader) loadGlobalConfig(configFile string) (PropertyConfigurationFile, error) {
	var cfg PropertyConfigurationFile

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return cfg, errors.New("global configuration file not found")
	}

	log.Debug("Loading global configuration file " + configFile)
	configor.New(&configor.Config{Debug: false}).Load(&cfg, configFile)

	return cfg, nil
}

/**
 * Save the global config
 */
func (configurationLoader ConfigurationLoader) saveGlobalConfig(configFile string, cfg PropertyConfigurationFile) error {
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
	var projectDirectory = ""

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
