package main

import (
	"bytes"
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
	util "github.com/EnvCLI/EnvCLI/pkg/util"
	colorable "github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v2"
)

// App Properties
var appName = "EnvCLI Utility"
var appVersion = "v0.4.2"

// Configuration
var defaultConfigurationDirectory = util.GetExecutionDirectory()

// Constants
var isCIEnvironment = util.IsCIEnvironment()

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
		os.Setenv("HTTP_PROXY", config.GetOrDefault(propConfig.Properties, "http-proxy", ""))
		os.Setenv("HTTPS_PROXY", config.GetOrDefault(propConfig.Properties, "https-proxy", ""))

		// Initialize Analytics
		if config.GetOrDefault(propConfig.Properties, "analytics", "true") == "true" {
			analytic.InitializeAnalytics(appName, appVersion)
		}
	}

	// Update Check, once a day (not in CI)
	appUpdater := updater.ApplicationUpdater{BintrayOrg: "envcli", BintrayRepository: "golang", BintrayPackage: "envcli", GitHubOrg: "EnvCLI", GitHubRepository: "EnvCLI"}
	var lastUpdateCheck, _ = strconv.ParseInt(config.GetOrDefault(propConfig.Properties, "last-update-check", strconv.Itoa(int(time.Now().Unix()))), 10, 64)
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
		After: func(c *cli.Context) error {
			// cleanup
			analytic.CleanUp()

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

					// iterate and quote args if needed
					commandArgs := append([]string{commandName}, c.Args().Tail()...)
					var commandWithArguments bytes.Buffer
					for _, arg := range commandArgs {
						if strings.Contains(arg, " ") {
							i := strings.Index(arg, "=")
							if i > -1 {
								argName := arg[0:i]
								argValue := arg[i+1 : len(arg)]
								fullArg := strings.Replace(argName+"="+strconv.Quote(argValue), "\"", "\\\"", -1)

								// quote for powershell, differs from the quoting for unix-based systems
								if runtime.GOOS == "windows" {
									fullArg = strings.Replace(argName+"="+strconv.Quote(argValue), "\"", "`\"", -1)
								}

								commandWithArguments.WriteString(fullArg)
							} else {
								commandWithArguments.WriteString(arg)
							}
						} else {
							commandWithArguments.WriteString(arg)
						}

						commandWithArguments.WriteString(" ")
					}

					log.Debugf("Received request to run command [%s] - with Arguments [%s].", commandName, strings.TrimSpace(commandWithArguments.String()))

					// Tracking: command
					analytic.TriggerEvent("Run", commandName)

					// config: try to load command configuration
					commandConfig, commandConfigErr := config.GetCommandConfiguration(commandName, util.GetWorkingDirectory())
					if commandConfigErr != nil {
						log.Errorf(commandConfigErr.Error())
						sentry.HandleError(commandConfigErr)
						return nil
					}

					// feature: before_script
					var commandWithBeforeScript = ""
					commandWithBeforeScript = strings.TrimSpace(commandWithArguments.String())
					if commandConfig.BeforeScript != nil {
						commandWithBeforeScript = strings.Join(commandConfig.BeforeScript[:], ";") + " && " + commandWithBeforeScript

						commandWithBeforeScript = strings.Replace(commandWithBeforeScript, "{HTTPProxy}", config.GetOrDefault(propConfig.Properties, "http-proxy", ""), -1)
						commandWithBeforeScript = strings.Replace(commandWithBeforeScript, "{HTTPSProxy}", config.GetOrDefault(propConfig.Properties, "https-proxy", ""), -1)
					}

					// feature: project mount
					var containerMounts []docker.ContainerMount
					var projectOrExecutionDir = config.GetProjectOrWorkingDirectory()
					containerMounts = append(containerMounts, docker.ContainerMount{MountType: "directory", Source: projectOrExecutionDir, Target: commandConfig.Directory})

					// feature: caching
					for _, cachingEntry := range commandConfig.Caching {
						var cacheFolder = config.GetOrDefault(propConfig.Properties, "cache-path", "") + "/" + cachingEntry.Name
						util.CreateDirectory(cacheFolder)
						containerMounts = append(containerMounts, docker.ContainerMount{MountType: "directory", Source: config.GetOrDefault(propConfig.Properties, "cache-path", "") + "/" + cachingEntry.Name, Target: cachingEntry.ContainerDirectory})
					}

					// feature: pass environment variables
					var environmentVariables []string = c.StringSlice("env")

					// feature: pass all env variables (excludes system variables like PATH, ...) in CI environments
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

					// feature: proxy environment
					if propConfigErr == nil {
						httpProxy := config.GetOrDefault(propConfig.Properties, "http-proxy", "")
						if httpProxy != "" {
							environmentVariables = append(environmentVariables, `http_proxy=`+httpProxy)
						}

						httpsProxy := config.GetOrDefault(propConfig.Properties, "https-proxy", "")
						if httpsProxy != "" {
							environmentVariables = append(environmentVariables, `https_proxy=`+httpsProxy)
						}
					}

					// detect container service and send command
					log.Infof("Executing command in container [%s].", commandConfig.Image)
					docker.ContainerExec(commandConfig.Image, commandConfig.Entrypoint, commandConfig.Shell, commandWithBeforeScript, containerMounts, commandConfig.Directory+"/"+util.GetPathRelativeToDirectory(util.GetWorkingDirectory(), projectOrExecutionDir), environmentVariables, c.StringSlice("port"))

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
						var globalConfigPath = config.GetOrDefault(propConfig.Properties, "global-configuration-path", defaultConfigurationDirectory)
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
					analytic.TriggerEvent("Aliases", scopeFilter)

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
							log.Infof("%s [%s]", varName, config.GetPropertyConfigEntry(varName))

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
