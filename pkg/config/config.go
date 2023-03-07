package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/cidverse/cidverseutils/pkg/collection"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/jinzhu/configor"
	"github.com/rs/zerolog/log"

	"gopkg.in/yaml.v2"
)

// Configuration
var defaultConfigurationDirectory = filesystem.GetExecutionDirectory()
var defaultConfigurationFile = ".envclirc"

// Constants
var validConfigurationOptions = []string{"http-proxy", "https-proxy", "global-configuration-path", "cache-path", "last-update-check"}

// LoadProjectConfig loads the project configuration
func LoadProjectConfig(configFile string) (ConfigurationFile, error) {
	log.Debug().Msg("Loading project configuration file " + configFile)
	var cfg ConfigurationFile

	file, err := os.Open(configFile)
	if err != nil {
		return ConfigurationFile{}, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return ConfigurationFile{}, err
	}

	return cfg, nil
}

// LoadPropertyConfig loads the property data
func LoadPropertyConfig() (PropertyConfigurationFile, error) {
	propConfigFile := defaultConfigurationDirectory + "/" + defaultConfigurationFile

	if _, err := os.Stat(propConfigFile); err == nil {
		return LoadPropertyConfigFile(defaultConfigurationDirectory + "/" + defaultConfigurationFile)
	}

	return PropertyConfigurationFile{}, nil
}

// LoadPropertyConfigFile loads the property config file
func LoadPropertyConfigFile(configFile string) (PropertyConfigurationFile, error) {
	log.Debug().Msg("Loading property configuration file " + configFile)
	var cfg PropertyConfigurationFile
	cfg.Properties = make(map[string]string)

	configor.New(&configor.Config{Debug: false}).Load(&cfg, configFile)

	return cfg, nil
}

// SavePropertyConfig saves the global config
func SavePropertyConfig(cfg PropertyConfigurationFile) error {
	return SavePropertyConfigFile(defaultConfigurationDirectory+"/"+defaultConfigurationFile, cfg)
}

// SavePropertyConfigFile saves the property file
func SavePropertyConfigFile(configFile string, cfg PropertyConfigurationFile) error {
	log.Debug().Msg("Saving property configuration file " + configFile)

	fileContent, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, fileContent, 0600)
}

// SetPropertyConfigEntry sets a property in the property config
func SetPropertyConfigEntry(varName string, varValue string) {
	// Load Config
	propConfig, _ := LoadPropertyConfig()

	// Set value
	isValidValue, _ := collection.InArray(varName, validConfigurationOptions)
	if isValidValue {
		propConfig.Properties[varName] = varValue

		// Save Config
		SavePropertyConfig(propConfig)
	}
}

// GetPropertyConfigEntry gets a property from the property config
func GetPropertyConfigEntry(varName string) string {
	// Load Config
	propConfig, _ := LoadPropertyConfig()

	// Get Value
	isValidValue, _ := collection.InArray(varName, validConfigurationOptions)
	if isValidValue {
		return propConfig.Properties[varName]
	}

	return ""
}

// UnsetPropertyConfigEntry clears a property
func UnsetPropertyConfigEntry(varName string) {
	// Load Config
	propConfig, _ := LoadPropertyConfig()

	// Get Value
	isValidValue, _ := collection.InArray(varName, validConfigurationOptions)
	if isValidValue {
		propConfig.Properties[varName] = ""

		// Save Config
		SavePropertyConfig(propConfig)
	}
}

// GetProjectOrWorkingDirectory returns either the project directory, if one can be found or the working directory
func GetProjectOrWorkingDirectory() string {
	var directory, err = GetProjectDirectory()
	if err != nil {
		directory = filesystem.GetWorkingDirectory()
	}
	return directory
}

// GetProjectDirectory searches for the project root directory by looking for the envcli config
func GetProjectDirectory() (string, error) {
	log.Trace().Msg("Trying to detect project directory ...")

	currentDirectory := filesystem.GetWorkingDirectory()
	var projectDirectory = ""
	log.Trace().Str("dir", currentDirectory).Msg("current working directory")

	directoryParts := strings.Split(currentDirectory, string(os.PathSeparator))

	for projectDirectory == "" {
		if _, err := os.Stat(filepath.Join(currentDirectory, "/.envcli.yml")); err == nil {
			log.Debug().Str("dir", currentDirectory).Msg("found project config in directory")
			return currentDirectory, nil
		}

		if directoryParts[0]+"\\" == currentDirectory || currentDirectory == "/" {
			log.Debug().Msg("didn't find a envcli project config in any parent directories")
			return "", errors.New("didn't find a envcli project config in any parent directories")
		}

		currentDirectory = filepath.Dir(currentDirectory)
		log.Trace().Str("dir", currentDirectory).Msg("proceed to search next directory")
	}

	return "", errors.New("didn't find a envcli project config in any parent directories")
}

// MergeConfigurations merges two configurations and keep the origin in the scope
func MergeConfigurations(configProject ConfigurationFile, configGlobal ConfigurationFile) ConfigurationFile {
	var cfg = ConfigurationFile{}

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

// GetCommandConfiguration gets the configuration entry for a specified command in the specified directory
func GetCommandConfiguration(commandName string, currentDirectory string, customIncludes []string) (RunConfigurationEntry, error) {
	// Global Configuration
	propConfig, propConfigErr := LoadPropertyConfig()
	if propConfigErr != nil {
		// error, when loading the config
		var emptyEntry RunConfigurationEntry
		return emptyEntry, propConfigErr
	}

	// Configuration file list
	var configFiles []string
	// - project directory
	projectDir, projectDirErr := GetProjectDirectory()
	if projectDirErr == nil {
		log.Debug().Msg("Project Directory: " + projectDir)
		configFiles = append(configFiles, projectDir+"/.envcli.yml")
	}
	// - custom includes
	configFiles = append(configFiles, customIncludes...)
	// - global (user-scope) configuration
	var globalConfigPath = collection.MapGetValueOrDefault(propConfig.Properties, "global-configuration-path", defaultConfigurationDirectory)
	log.Debug().Msg("Will load the global configuration from " + globalConfigPath + ".")
	configFiles = append(configFiles, globalConfigPath+"/.envcli.yml")

	// load configuration files
	var finalConfiguration ConfigurationFile
	for _, configFile := range configFiles {
		configContent, _ := LoadProjectConfig(configFile)
		finalConfiguration = MergeConfigurations(finalConfiguration, configContent)
	}

	// search for command definition
	for _, element := range finalConfiguration.Images {
		log.Debug().Msg("Checking for a match in image " + element.Name + " [Scope: " + element.Scope + "]")
		for _, providedCommand := range element.Provides {
			if providedCommand == commandName {
				log.Debug().Msg("Matched command " + commandName + " in package [" + element.Name + "]")

				return element, nil
			}
		}
	}

	// didn't find a match, error
	var emptyEntry RunConfigurationEntry
	return emptyEntry, errors.New("no configuration for command " + commandName + " found")
}
