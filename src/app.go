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
		os.Setenv("HTTP_PROXY", propConfig.HTTPProxy)
		os.Setenv("HTTPS_PROXY", propConfig.HTTPSProxy)
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

					// load global configuration
					var globalConfigPath = defaultConfigurationDirectory
					if propConfigErr == nil && propConfig.ConfigurationPath != "" {
						globalConfigPath = propConfig.ConfigurationPath
					}
					log.Warnf("[%s].", globalConfigPath)
					globalConfig, _ := configurationLoader.loadProjectConfig(globalConfigPath + "/.envcli.yml")

					// load project configuration
					configurationLoader := ConfigurationLoader{}
					if configurationLoader.getProjectDirectory() == "" {
						log.Warnf("No project configuration found in current or parent directories. Only the globally defined commands are available.")
						return nil
					}
					projectConfig, _ := configurationLoader.loadProjectConfig(configurationLoader.getProjectDirectory() + "/.envcli.yml")

					// merge project and global configuration
					var finalConfiguration = configurationLoader.mergeConfigurations(projectConfig, globalConfig)

					// check for command prefix and get the matching configuration entry
					var dockerImage = ""
					var dockerImageTag = ""
					var projectDirectory = ""
					var commandShell = ""
					var commandWithBeforeScript = ""
					for _, element := range finalConfiguration.Commands {
						log.Debugf("Checking for matching commands in package %s [Scope: %s]", element.Name, element.Scope)
						for _, providedCommand := range element.Provides {
							if providedCommand == commandName {
								log.Debugf("Matched command %s in package [%s]", commandName, element.Name)
								dockerImage = element.Image
								dockerImageTag = element.Tag
								projectDirectory = element.Directory
								commandShell = element.Shell

								commandWithBeforeScript = commandWithArguments
								if element.BeforeScript != nil {
									commandWithBeforeScript = strings.Join(element.BeforeScript[:], ";") + " && " + commandWithArguments

									commandWithBeforeScript = strings.Replace(commandWithBeforeScript, "{HTTPProxy}", propConfig.HTTPProxy, -1)
									commandWithBeforeScript = strings.Replace(commandWithBeforeScript, "{HTTPSProxy}", propConfig.HTTPSProxy, -1)
								}

								log.Debugf("Image: %s | Tag: %s | ImageDirectory: %s", dockerImage, dockerImageTag, projectDirectory)
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
						if propConfig.HTTPProxy != "" {
							environmentVariables = append(environmentVariables, "http_proxy="+propConfig.HTTPProxy)
						}
						if propConfig.HTTPSProxy != "" {
							environmentVariables = append(environmentVariables, "https_proxy="+propConfig.HTTPSProxy)
						}
					}

					// detect container service and send command
					log.Infof("Executing command in container [%s:%s].", dockerImage, dockerImageTag)
					docker := Docker{}
					docker.containerExec(dockerImage, dockerImageTag, commandShell, commandWithBeforeScript, configurationLoader.getProjectDirectory(), projectDirectory, projectDirectory+"/"+configurationLoader.getRelativePathToWorkingDirectory(), environmentVariables, c.StringSlice("port"))

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

							if varName == "http-proxy" {
								propConfig.HTTPProxy = varValue
								log.Infof("Set value of %s to [%s]", varName, propConfig.HTTPProxy)
							} else if varName == "https-proxy" {
								propConfig.HTTPSProxy = varValue
								log.Infof("Set value of %s to [%s]", varName, propConfig.HTTPSProxy)
							} else if varName == "configuration-path" {
								propConfig.ConfigurationPath = varValue
								log.Infof("Set value of %s to [%s]", varName, propConfig.ConfigurationPath)
							} else {
								log.Infof("Unknown variable name [%s]", varName)
							}

							// Save Config
							configurationLoader.savePropertyConfig(defaultConfigurationDirectory+"/.envclirc", propConfig)

							return nil
						},
					},
					&cli.Command{
						Name: "get",
						Action: func(c *cli.Context) error {
							// Check Parameters
							if c.NArg() != 1 {
								log.Fatal("Please provide the variable name you want to erase. [envcli config unset variable]")
							}
							varName := c.Args().Get(0)

							// Get Value
							if varName == "http-proxy" {
								log.Infof("%s [%s]", varName, propConfig.HTTPProxy)
							} else if varName == "https-proxy" {
								log.Infof("%s [%s]", varName, propConfig.HTTPSProxy)
							} else if varName == "configuration-path" {
								log.Infof("%s [%s]", varName, propConfig.ConfigurationPath)
							} else {
								log.Infof("Unknown variable name [%s]", varName)
							}

							return nil
						},
					},
					&cli.Command{
						Name: "unset",
						Action: func(c *cli.Context) error {
							// Check Parameters
							if c.NArg() != 1 {
								log.Fatal("Please provide the variable name you want to read. [envcli config get variable]")
							}
							varName := c.Args().Get(0)

							// Get Value
							if varName == "http-proxy" {
								propConfig.HTTPProxy = ""
							} else if varName == "https-proxy" {
								propConfig.HTTPSProxy = ""
							} else if varName == "configuration-path" {
								propConfig.ConfigurationPath = ""
							} else {
								log.Fatalf("Unknown variable name [%s]", varName)
								return nil
							}

							log.Infof("Unset variable %s.", varName)

							// Save Config
							configurationLoader.savePropertyConfig(defaultConfigurationDirectory+"/.envclirc", propConfig)

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
