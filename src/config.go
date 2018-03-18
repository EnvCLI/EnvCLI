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
	Commands []struct {
		// name of the container
		Name string

		// description for the  container
		Description string

		// the commands provided by the image
		Provides []string

		// docker image
		Image string

		// docker image tag
		Tag string

		// target directory to mount your project inside of the container
		Directory string `default:"/project"`

		// wrap the executed command inside of the container into a shell (ex. if you use globs)
		Shell string `default:"none"`

		// the command scope (internal use only) - global or project
		Scope string
	}
}

/**
 * The EnvCLI Configuration
 */
type PropertyConfigurationFile struct {
	ConfigurationPath string `default:""`
	HTTPProxy         string `default:""`
	HTTPSProxy        string `default:""`
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
 * Load the property config
 */
func (configurationLoader ConfigurationLoader) loadPropertyConfig(configFile string) (PropertyConfigurationFile, error) {
	var cfg PropertyConfigurationFile

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return cfg, errors.New("global property file not found")
	}

	log.Debug("Loading property configuration file " + configFile)
	configor.New(&configor.Config{Debug: false}).Load(&cfg, configFile)

	return cfg, nil
}

/**
 * Save the global config
 */
func (configurationLoader ConfigurationLoader) savePropertyConfig(configFile string, cfg PropertyConfigurationFile) error {
	log.Debug("Saving property configuration file " + configFile)

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
		}

		if directoryParts[0]+"\\" == currentDirectory {
			return ""
		}

		currentDirectory = filepath.Dir(currentDirectory)
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

/**
 * Merge two configurations and keep the origin in the Scope
 * TODO: Handle conflicts with a warning / by order project definition have precedence right now
 */
func (configurationLoader ConfigurationLoader) mergeConfigurations(configProject ProjectConfigrationFile, configGlobal ProjectConfigrationFile) ProjectConfigrationFile {
	var cfg = ProjectConfigrationFile{}

	for _, command := range configProject.Commands {
		command.Scope = "Project"
		cfg.Commands = append(cfg.Commands, command)
	}
	for _, command := range configGlobal.Commands {
		command.Scope = "Global"
		cfg.Commands = append(cfg.Commands, command)
	}

	return cfg
}
