package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/jinzhu/configor"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

// Configuration
var defaultConfigurationDirectory = GetExecutionDirectory()
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
 * Get the project root directory by searching for the envcli config
 */
func GetProjectDirectory() string {
	log.WithFields(log.Fields{
		"method": "getProjectDirectory()",
	}).Debugf("Trying to detect project directory ...")

	currentDirectory := GetWorkingDirectory()
	var projectDirectory = ""
	log.WithFields(log.Fields{
		"method": "getProjectDirectory()",
	}).Debugf("current working directory [%s]", currentDirectory)

	directoryParts := strings.Split(currentDirectory, string(os.PathSeparator))

	for projectDirectory == "" {
		if _, err := os.Stat(filepath.Join(currentDirectory, "/.envcli.yml")); err == nil {
			log.WithFields(log.Fields{
				"method": "getProjectDirectory()",
			}).Debugf("found project config in directory [%s]", currentDirectory)
			return currentDirectory
		}

		if directoryParts[0]+"\\" == currentDirectory || currentDirectory == "/" {
			log.WithFields(log.Fields{
				"method": "getProjectDirectory()",
			}).Debugf("didn't find a envcli project config in any parent directors")
			return ""
		}

		currentDirectory = filepath.Dir(currentDirectory)
		log.WithFields(log.Fields{
			"method": "getProjectDirectory()",
		}).Debugf("proceed to search next directory [%s]", currentDirectory)
	}

	return ""
}

/**
 * Merge two configurations and keep the origin in the Scope
 * TODO: Handle conflicts with a warning / by order project definition have precedence right now
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
