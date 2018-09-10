package main

import (
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	colorable "github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v2"
)

// App Properties
var appName = "EnvCLI Utility"
var appVersion = "v0.2.0"

// Configuration
var configurationLoader = ConfigurationLoader{}
var defaultConfigurationDirectory = configurationLoader.getExecutionDirectory()

// Constants
var validConfigurationOptions = []string{"http-proxy", "https-proxy", "global-configuration-path", "cache-path"}

// Init Hook
func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	// Fix color output for windows [https://github.com/Sirupsen/logrus/issues/172]
	if runtime.GOOS == "windows" {
		log.SetFormatter(&log.TextFormatter{ForceColors: true})
		log.SetOutput(colorable.NewColorableStdout())
	}
}

// CLI Main Entrypoint
func main() {

	// Global Configuration
	propConfig, propConfigErr := configurationLoader.loadPropertyConfig(defaultConfigurationDirectory + "/.envclirc")

	// Configure Proxy Server
	if propConfigErr == nil {
		// Set Proxy Server
		os.Setenv("HTTP_PROXY", getOrDefault(propConfig.Properties, "http-proxy", ""))
		os.Setenv("HTTPS_PROXY", getOrDefault(propConfig.Properties, "https-proxy", ""))
	}

	// CLI
	app := &cli.App{
		Name:                  appName,
		Version:               appVersion,
		Compiled:              time.Now(),
		EnableShellCompletion: true,
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Philipp Heuer",
				Email: "git@philippheuer.me",
			},
		},
		Usage: "Runs cli commands within docker containers to provide a modern development environment",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "loglevel",
				Value: "info",
				Usage: "The loglevel used by envcli, use this to troubleshoot issues",
			},
		},
		Before: func(c *cli.Context) error {
			// Set loglevel
			setLoglevel(c.String("loglevel"))

			return nil
		},
		Commands: []*cli.Command{
			/**
			 * Command: self-update
			 */
			{
				Name:    "self-update",
				Aliases: []string{},
				Usage:   "updates the dev cli utility",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "force",
						Aliases: []string{"f"},
						Value:   false,
						Usage:   "A forced update would also redownload the current version",
					},
				},
				Action: func(c *cli.Context) error {
					// Run Update
					appUpdater := ApplicationUpdater{BintrayOrg: "envcli", BintrayRepository: "golang", BintrayPackage: "envcli", GitHubOrg: "EnvCLI", GitHubRepository: "EnvCLI"}
					appUpdater.update("latest", c.Bool("force"))

					return nil
				},
			},
			/**
			 * Command: run
			 */
			{
				Name:    "run",
				Aliases: []string{},
				Usage:   "runs 3rd party commands within their respective docker containers",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:    "env",
						Aliases: []string{"e"},
						Usage:   "Sets environment variables within the containers",
					},
					&cli.StringSliceFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Usage:   "Publish ports of the container",
					},
				},
				Action: func(c *cli.Context) error {
					// parse command
					commandName := c.Args().First()
					commandWithArguments := strings.Join(append([]string{commandName}, c.Args().Tail()...), " ")
					log.Debugf("Received request to run command [%s] - with Arguments [%s].", commandName, commandWithArguments)

					// load global (user-scope) configuration
					var globalConfigPath = getOrDefault(propConfig.Properties, "global-configuration-path", defaultConfigurationDirectory)
					log.Debugf("Will load the global configuration from [%s].", globalConfigPath)
					globalConfig, _ := configurationLoader.loadProjectConfig(globalConfigPath + "/.envcli.yml")

					// load project configuration
					configurationLoader := ConfigurationLoader{}
					var projectDirectory = configurationLoader.getProjectDirectory()
					if projectDirectory == "" {
						log.Warnf("No project configuration found in current or parent directories. Only the global commands are available.")
						projectDirectory = getWorkingDirectory()
					}
					log.Debugf("Project Directory: %s", projectDirectory)
					projectConfig, _ := configurationLoader.loadProjectConfig(projectDirectory + "/.envcli.yml")

					// merge project and global configuration
					var finalConfiguration = configurationLoader.mergeConfigurations(projectConfig, globalConfig)

					// check for command prefix and get the matching configuration entry
					var dockerImage = ""
					var containerDirectory = ""
					var commandShell = ""
					var commandWithBeforeScript = ""
					var containerMounts []ContainerMount

					for _, element := range finalConfiguration.Images {
						log.Debugf("Checking for a match in image %s [Scope: %s]", element.Name, element.Scope)
						for _, providedCommand := range element.Provides {
							if providedCommand == commandName {
								log.Debugf("Matched command %s in package [%s]", commandName, element.Name)
								dockerImage = element.Image
								containerDirectory = element.Directory
								commandShell = element.Shell

								commandWithBeforeScript = commandWithArguments
								if element.BeforeScript != nil {
									commandWithBeforeScript = strings.Join(element.BeforeScript[:], ";") + " && " + commandWithArguments

									commandWithBeforeScript = strings.Replace(commandWithBeforeScript, "{HTTPProxy}", getOrDefault(propConfig.Properties, "http-proxy", ""), -1)
									commandWithBeforeScript = strings.Replace(commandWithBeforeScript, "{HTTPSProxy}", getOrDefault(propConfig.Properties, "https-proxy", ""), -1)
								}

								// project mount
								containerMounts = append(containerMounts, ContainerMount{mountType: "directory", source: projectDirectory, target: containerDirectory})

								// caching mounts
								for _, cachingEntry := range element.Caching {
									var cacheFolder = getOrDefault(propConfig.Properties, "cache-path", "") + "/" + cachingEntry.Name
									createDirectory(cacheFolder)
									containerMounts = append(containerMounts, ContainerMount{mountType: "directory", source: getOrDefault(propConfig.Properties, "cache-path", "") + "/" + cachingEntry.Name, target: cachingEntry.ContainerDirectory})
								}

								log.Debugf("Image: %s | ImageDirectory: %s", dockerImage, containerDirectory)
							}
						}
					}
					if dockerImage == "" {
						log.Errorf("No configuration for command [%s] found.", commandName)
						return nil
					}

					// environment variables
					var environmentVariables []string = c.StringSlice("env")

					// - proxy environment
					if propConfigErr == nil {
						httpProxy := getOrDefault(propConfig.Properties, "http-proxy", "")
						if httpProxy != "" {
							environmentVariables = append(environmentVariables, "http_proxy="+httpProxy)
						}

						httpsProxy := getOrDefault(propConfig.Properties, "https-proxy", "")
						if httpsProxy != "" {
							environmentVariables = append(environmentVariables, "https_proxy="+httpsProxy)
						}
					}

					// detect container service and send command
					log.Infof("Executing command in container [%s].", dockerImage)
					docker := Docker{}
					docker.containerExec(dockerImage, commandShell, commandWithBeforeScript, containerMounts, containerDirectory+"/"+getPathRelativeToDirectory(getWorkingDirectory(), projectDirectory), environmentVariables, c.StringSlice("port"))

					return nil
				},
			},
			/**
			 * Command: config
			 */
			{
				Name:    "config",
				Aliases: []string{},
				Usage:   "updates the dev cli utility",
				Subcommands: []*cli.Command{
					&cli.Command{
						Name: "set",
						Action: func(c *cli.Context) error {
							// Load Config
							configurationLoader := ConfigurationLoader{}
							propConfig, _ := configurationLoader.loadPropertyConfig(defaultConfigurationDirectory + "/.envclirc")

							// Check Parameters
							if c.NArg() != 2 {
								log.Fatal("Please provide the variable name and the value you want to set in this format. [envcli config set variable value]")
							}
							varName := c.Args().Get(0)
							varValue := c.Args().Get(1)

							// Set value
							isValidValue, _ := inArray(varName, validConfigurationOptions)
							if isValidValue {
								propConfig.Properties[varName] = varValue
								log.Infof("Set value of %s to [%s]", varName, varValue)

								// Save Config
								configurationLoader.savePropertyConfig(defaultConfigurationDirectory+"/.envclirc", propConfig)
							} else {
								log.Warnf("Unknown variable [%s]", varName)
							}

							return nil
						},
					},
					&cli.Command{
						Name: "get",
						Action: func(c *cli.Context) error {
							// Check Parameters
							if c.NArg() != 1 {
								log.Fatal("Please provide the variable name you want to read. [envcli config get variable]")
							}
							varName := c.Args().Get(0)

							// Get Value
							isValidValue, _ := inArray(varName, validConfigurationOptions)
							if isValidValue {
								log.Infof("%s [%s]", propConfig.Properties[varName])
							} else {
								log.Warnf("Unknown variable [%s]", varName)
							}

							return nil
						},
					},
					&cli.Command{
						Name: "get-all",
						Action: func(c *cli.Context) error {
							// Print all values
							for key, value := range propConfig.Properties {
								log.Infof("%s [%s]", key, value)
							}

							return nil
						},
					},
					&cli.Command{
						Name: "unset",
						Action: func(c *cli.Context) error {
							// Check Parameters
							if c.NArg() != 1 {
								log.Fatal("Please provide the variable name you want to unset. [envcli config unset variable]")
							}
							varName := c.Args().Get(0)

							// Unset Value
							isValidValue, _ := inArray(varName, validConfigurationOptions)
							if isValidValue {
								propConfig.Properties[varName] = ""
								log.Infof("Value of variable %s set to [].", varName)

								// Save Config
								configurationLoader.savePropertyConfig(defaultConfigurationDirectory+"/.envclirc", propConfig)
							} else {
								log.Warnf("Unknown variable [%s]", varName)
							}

							return nil
						},
					},
				},
			},
		},
	}

	// Sort Flags & Commands by Alphabet
	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	// Run Application
	app.Run(os.Args)
}
