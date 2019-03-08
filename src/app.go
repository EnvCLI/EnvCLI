package main

import (
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	aliases "github.com/EnvCLI/EnvCLI/pkg/aliases"
	analytic "github.com/EnvCLI/EnvCLI/pkg/analytic"
	config "github.com/EnvCLI/EnvCLI/pkg/config"
	docker "github.com/EnvCLI/EnvCLI/pkg/docker"
	sentry "github.com/EnvCLI/EnvCLI/pkg/sentry"
	updater "github.com/EnvCLI/EnvCLI/pkg/updater"
	colorable "github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v2"
)

// App Properties
var appName = "EnvCLI Utility"
var appVersion = "v0.4.0"

// Configuration
var defaultConfigurationDirectory = config.GetExecutionDirectory()

// Constants
var isCIEnvironment = DetectCIEnvironment()

// Init Hook
func init() {
	// Initialize SentryIO
	sentry.InitializeSentryIO(appVersion)

	// Logging
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
	propConfig, propConfigErr := config.LoadPropertyConfig()

	// Configure Proxy Server
	if propConfigErr == nil {
		// Set Proxy Server
		os.Setenv("HTTP_PROXY", getOrDefault(propConfig.Properties, "http-proxy", ""))
		os.Setenv("HTTPS_PROXY", getOrDefault(propConfig.Properties, "https-proxy", ""))

		// Initialize Analytics
		if getOrDefault(propConfig.Properties, "analytics", "true") == "true" {
			analytic.InitializeAnalytics(appName, appName)
		}
	}

	// Tracking
	analytic.TriggerEvent("OS", runtime.GOOS)
	analytic.TriggerEvent("Version", appVersion)
	if isCIEnvironment {
		analytic.TriggerEvent("Platform", "CI")
	} else {
		analytic.TriggerEvent("Platform", "DESKTOP")
	}

	// Update Check, once a day (not in CI)
	appUpdater := updater.ApplicationUpdater{BintrayOrg: "envcli", BintrayRepository: "golang", BintrayPackage: "envcli", GitHubOrg: "EnvCLI", GitHubRepository: "EnvCLI"}
	var lastUpdateCheck, _ = strconv.ParseInt(getOrDefault(propConfig.Properties, "last-update-check", strconv.Itoa(int(time.Now().Unix()))), 10, 64)
	if time.Now().Unix() >= lastUpdateCheck+86400 && isCIEnvironment == false {
		if appUpdater.IsUpdateAvailable(appVersion) {
			log.Warnf("You are using a old version, please consider to update using `envcli self-update`!")
		}
	}
	if isCIEnvironment == false {
		config.SetPropertyConfigEntry("last-update-check", strconv.Itoa(int(time.Now().Unix())))
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
					appUpdater.Update("latest", c.Bool("force"), appVersion)

					// Tracking: Command
					analytic.TriggerEvent("Update", "Execute")

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

					// Tracking: Command
					analytic.TriggerEvent("CommandExecution", commandName)

					// load global (user-scope) configuration
					var globalConfigPath = getOrDefault(propConfig.Properties, "global-configuration-path", defaultConfigurationDirectory)
					log.Debugf("Will load the global configuration from [%s].", globalConfigPath)
					globalConfig, _ := config.LoadProjectConfig(globalConfigPath + "/.envcli.yml")

					// load project configuration
					var projectDirectory = config.GetProjectDirectory()
					if projectDirectory == "" {
						log.Warnf("No project configuration found in current or parent directories. Only the global commands are available.")
						projectDirectory = config.GetWorkingDirectory()
					}
					log.Debugf("Project Directory: %s", projectDirectory)
					projectConfig, _ := config.LoadProjectConfig(projectDirectory + "/.envcli.yml")

					// merge project and global configuration
					var finalConfiguration = config.MergeConfigurations(projectConfig, globalConfig)

					// check for command prefix and get the matching configuration entry
					var dockerImage = ""
					var containerDirectory = ""
					var entrypoint = ""
					var commandShell = ""
					var commandWithBeforeScript = ""
					var containerMounts []docker.ContainerMount

				configLoop:
					for _, element := range finalConfiguration.Images {
						log.Debugf("Checking for a match in image %s [Scope: %s]", element.Name, element.Scope)
						for _, providedCommand := range element.Provides {
							if providedCommand == commandName {
								log.Debugf("Matched command %s in package [%s]", commandName, element.Name)
								dockerImage = element.Image
								containerDirectory = element.Directory
								entrypoint = element.Entrypoint
								commandShell = element.Shell

								commandWithBeforeScript = commandWithArguments
								if element.BeforeScript != nil {
									commandWithBeforeScript = strings.Join(element.BeforeScript[:], ";") + " && " + commandWithArguments

									commandWithBeforeScript = strings.Replace(commandWithBeforeScript, "{HTTPProxy}", getOrDefault(propConfig.Properties, "http-proxy", ""), -1)
									commandWithBeforeScript = strings.Replace(commandWithBeforeScript, "{HTTPSProxy}", getOrDefault(propConfig.Properties, "https-proxy", ""), -1)
								}

								// project mount
								containerMounts = append(containerMounts, docker.ContainerMount{MountType: "directory", Source: projectDirectory, Target: containerDirectory})

								// caching mounts
								for _, cachingEntry := range element.Caching {
									var cacheFolder = getOrDefault(propConfig.Properties, "cache-path", "") + "/" + cachingEntry.Name
									createDirectory(cacheFolder)
									containerMounts = append(containerMounts, docker.ContainerMount{MountType: "directory", Source: getOrDefault(propConfig.Properties, "cache-path", "") + "/" + cachingEntry.Name, Target: cachingEntry.ContainerDirectory})
								}

								log.Debugf("Image: %s | ImageDirectory: %s", dockerImage, containerDirectory)
								break configLoop
							}
						}
					}
					if dockerImage == "" {
						log.Errorf("No configuration for command [%s] found.", commandName)
						return nil
					}

					// environment variables
					var environmentVariables []string = c.StringSlice("env")

					// auto provide ci env variables (excludes system variables like PATH, ...)
					if isCIEnvironment {
						for _, e := range os.Environ() {
							pair := strings.SplitN(e, "=", 2)
							var envName = pair[0]
							var envValue = pair[1]

							// filter vars
							var systemVars = []string{"_", "PWD", "OLDPWD", "PATH", "HOME", "HOSTNAME", "TERM", "SHLVL", "HTTP_PROXY", "HTTPS_PROXY"}
							isExluded, _ := config.InArray(strings.ToUpper(envName), systemVars)
							if !isExluded {
								log.Debugf("Added environment variable %s [%s] from host!", envName, envValue)
								environmentVariables = append(environmentVariables, envName+`=`+envValue)
							} else {
								log.Debugf("Excluded env variable %s [%s] from host based on the filter rule.", envName, envValue)
							}
						}
					}

					// - proxy environment
					if propConfigErr == nil {
						httpProxy := getOrDefault(propConfig.Properties, "http-proxy", "")
						if httpProxy != "" {
							environmentVariables = append(environmentVariables, `http_proxy=`+httpProxy)
						}

						httpsProxy := getOrDefault(propConfig.Properties, "https-proxy", "")
						if httpsProxy != "" {
							environmentVariables = append(environmentVariables, `https_proxy=`+httpsProxy)
						}
					}

					// detect container service and send command
					log.Infof("Executing command in container [%s].", dockerImage)
					docker.ContainerExec(dockerImage, entrypoint, commandShell, commandWithBeforeScript, containerMounts, containerDirectory+"/"+getPathRelativeToDirectory(getWorkingDirectory(), projectDirectory), environmentVariables, c.StringSlice("port"))

					return nil
				},
			},
			/**
			 * Command: install-aliases
			 */
			{
				Name:    "install-aliases",
				Aliases: []string{},
				Usage:   "installs aliases for the global / project scoped commands",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "scope",
						Aliases: []string{"s"},
						Value:   "all",
						Usage:   "Install aliases for the specified scope (project/global or all)",
					},
				},
				Action: func(c *cli.Context) error {
					log.Debugf("Installing aliases ...")
					scopeFilter := c.String("scope")

					// create global-scoped aliases
					if scopeFilter == "all" || scopeFilter == "global" {
						var globalConfigPath = getOrDefault(propConfig.Properties, "global-configuration-path", defaultConfigurationDirectory)
						log.Debugf("Will load the global configuration from [%s].", globalConfigPath)
						globalConfig, _ := config.LoadProjectConfig(globalConfigPath + "/.envcli.yml")

						for _, element := range globalConfig.Images {
							element.Scope = "Global"
							log.Debugf("Created aliases for %s [Scope: %s]", element.Name, element.Scope)

							// for each provided command
							for _, currentCommand := range element.Provides {
								aliases.InstallAlias(appVersion, currentCommand, element.Scope)
							}
						}
					}

					// create project-scoped aliases
					if scopeFilter == "all" || scopeFilter == "project" {
						var projectDirectory = config.GetProjectDirectory()
						log.Debugf("Project Directory: %s", projectDirectory)
						projectConfig, _ := config.LoadProjectConfig(projectDirectory + "/.envcli.yml")

						for _, element := range projectConfig.Images {
							element.Scope = "Project"
							log.Debugf("Created aliases for %s [Scope: %s]", element.Name, element.Scope)

							// for each provided command
							for _, currentCommand := range element.Provides {
								aliases.InstallAlias(appVersion, currentCommand, element.Scope)
							}
						}
					}

					// Tracking: Command
					analytic.TriggerEvent("Aliases", "Install")

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
							// Check Parameters
							if c.NArg() != 2 {
								log.Fatal("Please provide the variable name and the value you want to set in this format. [envcli config set variable value]")
							}
							varName := c.Args().Get(0)
							varValue := c.Args().Get(1)

							// Set value
							config.SetPropertyConfigEntry(varName, varValue)
							log.Infof("Set value of %s to [%s]", varName, varValue)

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
							log.Infof("%s [%s]", config.GetPropertyConfigEntry(varName))

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

							// Unset value
							config.UnsetPropertyConfigEntry(varName)
							log.Infof("Value of variable %s set to [].", varName)

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
