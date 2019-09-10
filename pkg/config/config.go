package config

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	util "github.com/EnvCLI/EnvCLI/pkg/util"
	"github.com/jinzhu/configor"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

// Configuration
var defaultConfigurationDirectory = util.GetExecutionDirectory()
var defaultConfigurationFile = ".envclirc"

// Constants
var validConfigurationOptions = []string{"http-proxy", "https-proxy", "global-configuration-path", "cache-path", "last-update-check"}

/**
 * Load the project config
 */
func LoadProjectConfig(configFile string) (ProjectConfigrationFile, error) {
	var cfg ProjectConfigrationFile

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Debugf("Can't load config - file [%s] does not exist!", configFile)
		return ProjectConfigrationFile{}, nil
	}

	log.Debugf("Loading project configuration file %s", configFile)
	configor.New(&configor.Config{Debug: false}).Load(&cfg, configFile)

	return cfg, nil
}

/**
 * Load the property config
 */
func LoadPropertyConfig() (PropertyConfigurationFile, error) {
	return LoadPropertyConfigFile(defaultConfigurationDirectory + "/" + defaultConfigurationFile)
}

/**
 * Load the property config file
 */
func LoadPropertyConfigFile(configFile string) (PropertyConfigurationFile, error) {
	var cfg PropertyConfigurationFile
	cfg.Properties = make(map[string]string)

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Debugf("Can't load global properties - file [%s] does not exist!", configFile)
		return cfg, nil
	}

	log.Debug("Loading property configuration file " + configFile)
	configor.New(&configor.Config{Debug: false}).Load(&cfg, configFile)

	return cfg, nil
}

/**
 * Save the global config
 */
func SavePropertyConfig(cfg PropertyConfigurationFile) error {
	return SavePropertyConfigFile(defaultConfigurationDirectory+"/"+defaultConfigurationFile, cfg)
}

/**
 * Save the global config file
 */
func SavePropertyConfigFile(configFile string, cfg PropertyConfigurationFile) error {
	log.Debug("Saving property configuration file " + configFile)

	fileContent, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(configFile, fileContent, 0600)
}

/**
 * Sets a property in the property config
 */
func SetPropertyConfigEntry(varName string, varValue string) {
	// Load Config
	propConfig, _ := LoadPropertyConfig()

	// Set value
	isValidValue, _ := InArray(varName, validConfigurationOptions)
	if isValidValue {
		propConfig.Properties[varName] = varValue

		// Save Config
		SavePropertyConfig(propConfig)
	}
}

/**
 * Gets a property in the property config
 */
func GetPropertyConfigEntry(varName string) string {
	// Load Config
	propConfig, _ := LoadPropertyConfig()

	// Get Value
	isValidValue, _ := InArray(varName, validConfigurationOptions)
	if isValidValue {
		return propConfig.Properties[varName]
	}

	return ""
}

/**
 * Gets a property in the property config
 */
func UnsetPropertyConfigEntry(varName string) {
	// Load Config
	propConfig, _ := LoadPropertyConfig()

	// Get Value
	isValidValue, _ := InArray(varName, validConfigurationOptions)
	if isValidValue {
		propConfig.Properties[varName] = ""

		// Save Config
		SavePropertyConfig(propConfig)
	}
}

/**
 * GetProjectOrWorkingDirectory returns either the project directory, if one can be found or the working directory
 */
func GetProjectOrWorkingDirectory() string {
	var directory, err = GetProjectDirectory()
	if err != nil {
		directory = util.GetWorkingDirectory()
	}
	return directory
}

/**
 * Get the project root directory by searching for the envcli config
 */
func GetProjectDirectory() (string, error) {
	log.Tracef("Trying to detect project directory ...")

	currentDirectory := GetWorkingDirectory()
	var projectDirectory = ""
	log.Tracef("current working directory [%s]", currentDirectory)

	directoryParts := strings.Split(currentDirectory, string(os.PathSeparator))

	for projectDirectory == "" {
		if _, err := os.Stat(filepath.Join(currentDirectory, "/.envcli.yml")); err == nil {
			log.Debugf("found project config in directory [%s]", currentDirectory)
			return currentDirectory, nil
		}

		if directoryParts[0]+"\\" == currentDirectory || currentDirectory == "/" {
			log.Debugf("didn't find a envcli project config in any parent directories")
			return "", errors.New("Didn't find a envcli project config in any parent directories")
		}

		currentDirectory = filepath.Dir(currentDirectory)
		log.Tracef("proceed to search next directory [%s]", currentDirectory)
	}

	return "", errors.New("Didn't find a envcli project config in any parent directories")
}

/**
 * Merge two configurations and keep the origin in the Scope
 */
func MergeConfigurations(configProject ProjectConfigrationFile, configGlobal ProjectConfigrationFile) ProjectConfigrationFile {
	var cfg = ProjectConfigrationFile{}

	for _, image := range configProject.Images {
		image.Scope = "Project"
		cfg.Images = append(cfg.Images, image)
	}
	for _, image := range configGlobal.Images {
		image.Scope = "Global"
		cfg.Images = append(cfg.Images, image)
	}

	return cfg
}

/**
 * GetCommandConfiguration gets the configuration entry for a specified command in the specified directory
 */
func GetCommandConfiguration(commandName string, currentDirectory string) (RunConfigurationEntry, error) {
	// Global Configuration
	propConfig, propConfigErr := LoadPropertyConfig()
	if propConfigErr != nil {
		// error, when loading the config
		var emptyEntry RunConfigurationEntry
		return emptyEntry, propConfigErr
	}

	// project directory
	projectDir, projectDirErr := GetProjectDirectory()

	// load global (user-scope) configuration
	var globalConfigPath = GetOrDefault(propConfig.Properties, "global-configuration-path", defaultConfigurationDirectory)
	log.Debugf("Will load the global configuration from [%s].", globalConfigPath)
	globalConfig, _ := LoadProjectConfig(globalConfigPath + "/.envcli.yml")

	// load project configuration
	var projectConfig ProjectConfigrationFile
	if projectDirErr == nil {
		log.Debugf("Project Directory: %s", projectDir)
		projectConfig, _ = LoadProjectConfig(projectDir + "/.envcli.yml")
	}

	// merge project and global configuration
	var finalConfiguration = MergeConfigurations(projectConfig, globalConfig)
	for _, element := range finalConfiguration.Images {
		log.Debugf("Checking for a match in image %s [Scope: %s]", element.Name, element.Scope)
		for _, providedCommand := range element.Provides {
			if providedCommand == commandName {
				log.Debugf("Matched command %s in package [%s]", commandName, element.Name)

				return element, nil
			}
		}
	}

	// didn't find a match, error
	var emptyEntry RunConfigurationEntry
	return emptyEntry, errors.New("no configuration for command " + commandName + " found")
}
