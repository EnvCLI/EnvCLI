package main

import (
	"fmt"
	"github.com/cidverse/cidverseutils/pkg/cihelper"
	"github.com/cidverse/cidverseutils/pkg/collection"
	"github.com/cidverse/cidverseutils/pkg/container_runtime"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	aliases "github.com/EnvCLI/EnvCLI/pkg/aliases"
	common "github.com/EnvCLI/EnvCLI/pkg/common"
	config "github.com/EnvCLI/EnvCLI/pkg/config"
	updater "github.com/EnvCLI/EnvCLI/pkg/updater"
	cli "github.com/urfave/cli/v2"
)

// Build Information
var (
	Version    = "dev"
	CommitHash = "none"
	BuildAt    = "unknown"
)

// Configuration
var defaultConfigurationDirectory = filesystem.GetExecutionDirectory()

// Constants
var isCIEnvironment = cihelper.IsCIEnvironment()

// Init Hook
func init() {
	// Initialize Global Logger
	colorableOutput := colorable.NewColorableStdout()
	log.Logger = zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: colorableOutput}).With().Timestamp().Logger()

	// Timestamp Format
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Only log the warning severity or above.
	zerolog.SetGlobalLevel(zerolog.WarnLevel)

	// detect debug mode
	debugValue, debugIsSet := os.LookupEnv("ENVCLI_DEBUG")
	if debugIsSet && strings.ToLower(debugValue) == "true" {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}

	// show calling files
	_, showCalls := os.LookupEnv("ENVCLI_SHOW_CALL")
	if showCalls {
		log.Logger = zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: colorableOutput}).With().Timestamp().Caller().Logger()
	}
}

// CLI Main Entrypoint
func main() {
	// Global Configuration
	propConfig, propConfigErr := config.LoadPropertyConfig()

	// Configure Proxy Server
	if propConfigErr == nil {
		// Set Proxy Server
		os.Setenv("HTTP_PROXY", collection.MapGetValueOrDefault(propConfig.Properties, "http-proxy", ""))
		os.Setenv("HTTPS_PROXY", collection.MapGetValueOrDefault(propConfig.Properties, "https-proxy", ""))
	}

	// Update Check, once a day (not in CI)
	appUpdater := updater.ApplicationUpdater{GitHubOrg: "EnvCLI", GitHubRepository: "EnvCLI"}
	var lastUpdateCheck, _ = strconv.ParseInt(collection.MapGetValueOrDefault(propConfig.Properties, "last-update-check", strconv.Itoa(int(time.Now().Unix()))), 10, 64)
	if time.Now().Unix() >= lastUpdateCheck+86400 && isCIEnvironment == false {
		if appUpdater.IsUpdateAvailable(Version) {
			log.Warn().Msg("You are using a old version, please consider to update using `envcli self-update`!")
		}
	}
	if isCIEnvironment == false {
		config.SetPropertyConfigEntry("last-update-check", strconv.Itoa(int(time.Now().Unix())))
	}

	// CLI
	app := &cli.App{
		Name:     "EnvCLI",
		Version:  Version,
		Compiled: time.Now(),
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
				Value: "warn",
				Usage: "The loglevel used by envcli, use this to troubleshoot issues",
			},
			&cli.StringSliceFlag{
				Name:    "config-include",
				Aliases: []string{},
				Usage:   "Additionally include these configuration files, please take note that precedence will be in this order: project config, included, system config",
			},
		},
		Before: func(c *cli.Context) error {
			// Set loglevel
			common.SetLoglevel(c.String("loglevel"))

			return nil
		},
		After: func(c *cli.Context) error {
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
						Usage:   "A forced update would also redownload the current version.",
					},
					&cli.StringFlag{
						Name:  "target",
						Value: "latest",
						Usage: "A target version that should be upgraded/downgraded to.",
					},
				},
				Action: func(c *cli.Context) error {
					// Run Update
					appUpdater.Update(c.String("target"), c.Bool("force"), Version)

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
					&cli.StringFlag{
						Name:    "userArgs",
						Aliases: []string{},
						Usage:   "Allows to specify custom arguments that will be passed to the docker run command for special cases",
					},
				},
				Action: func(c *cli.Context) error {
					// parse command
					commandName := c.Args().First()

					// iterate and quote args if needed
					commandArgs := append([]string{commandName}, c.Args().Tail()...)
					commandWithArguments := common.ParseAndEscapeArgs(commandArgs)

					log.Debug().Msg("Received request to run command [" + commandName + "] - with Arguments [" + commandWithArguments + "].")

					// config: try to load command configuration
					commandConfig, commandConfigErr := config.GetCommandConfiguration(commandName, filesystem.GetWorkingDirectory(), c.StringSlice("config-include"))
					if commandConfigErr != nil {
						log.Error().Err(commandConfigErr).Msg("failed to load command config")
						os.Exit(1)
						return nil
					}

					// container runtime
					containerRuntime := &container_runtime.ContainerRuntime{}
					container := containerRuntime.NewContainer()
					container.SetImage(commandConfig.Image)
					container.SetEntrypoint(commandConfig.Entrypoint)
					container.SetCommandShell(commandConfig.Shell)
					var projectOrExecutionDir = config.GetProjectOrWorkingDirectory()
					log.Debug().Str("source", projectOrExecutionDir).Str("target", commandConfig.Directory).Msg("Adding volume mount")
					container.AddVolume(container_runtime.ContainerMount{MountType: "directory", Source: projectOrExecutionDir, Target: commandConfig.Directory})
					container.SetWorkingDirectory(commandConfig.Directory + "/" + filesystem.GetPathRelativeToDirectory(filesystem.GetWorkingDirectory(), projectOrExecutionDir))

					// core: expose ports (command args)
					container.AddContainerPorts(c.StringSlice("port"))

					// core: pass environment variables (command args)
					container.AddEnvironmentVariables(c.StringSlice("env"))

					// feature: user args
					if len(c.String("userArgs")) > 0 {
						container.SetUserArgs(c.String("userArgs"))
					}

					// feature: before_script
					var commandWithBeforeScript = ""
					commandWithBeforeScript = strings.TrimSpace(commandWithArguments)
					if commandConfig.BeforeScript != nil {
						commandWithBeforeScript = strings.Join(commandConfig.BeforeScript[:], ";") + " && " + commandWithBeforeScript

						commandWithBeforeScript = strings.Replace(commandWithBeforeScript, "{HTTPProxy}", collection.MapGetValueOrDefault(propConfig.Properties, "http-proxy", ""), -1)
						commandWithBeforeScript = strings.Replace(commandWithBeforeScript, "{HTTPSProxy}", collection.MapGetValueOrDefault(propConfig.Properties, "https-proxy", ""), -1)
					}
					log.Debug().Msg("Setting new command with before_script: " + commandWithBeforeScript)
					container.SetCommand(commandWithBeforeScript)

					// feature: container runtime access
					if commandConfig.ContainerRuntimeAccess {
						container.AllowContainerRuntimeAcccess()
					}

					// feature: caching
					for _, cachingEntry := range commandConfig.Caching {
						if collection.MapGetValueOrDefault(propConfig.Properties, "cache-path", "") == "" {
							log.Warn().Msg("CachePath not set, not using the specified cache directories.")
							break
						}

						var cacheFolder = collection.MapGetValueOrDefault(propConfig.Properties, "cache-path", "") + "/" + cachingEntry.Name
						filesystem.CreateDirectory(cacheFolder)
						container.AddCacheMount(cachingEntry.Name, cacheFolder, cachingEntry.ContainerDirectory)
					}

					// feature: capabilities
					for _, cap := range commandConfig.CapAdd {
						container.AddCapability(cap)
					}

					// feature: pass all env variables (excludes system variables like PATH, ...) in CI environments
					if isCIEnvironment {
						container.AddAllEnvironmentVariables()
					}

					// feature: proxy environment
					if propConfigErr == nil {
						httpProxy := collection.MapGetValueOrDefault(propConfig.Properties, "http-proxy", "")
						if httpProxy != "" {
							container.AddEnvironmentVariable("http_proxy", httpProxy)
						}

						httpsProxy := collection.MapGetValueOrDefault(propConfig.Properties, "https-proxy", "")
						if httpsProxy != "" {
							container.AddEnvironmentVariable("https_proxy", httpsProxy)
						}
					}

					// detect container service and send command
					log.Info().Msg("Executing command in container [" + commandConfig.Image + "].")
					container.StartContainer()

					return nil
				},
			},
			/**
			 * Command: pull-image
			 */
			{
				Name:    "pull-image",
				Aliases: []string{},
				Usage:   "pulls the needed images for the specified commands",
				Action: func(c *cli.Context) error {
					commands := append([]string{c.Args().First()}, c.Args().Tail()...)

					// pull image for each provided command
					fmt.Printf("Pulling images for [%s].\n", strings.Join(commands, ", "))
					for _, cmd := range commands {
						log.Debug().Msg("Pulling image for command [" + cmd + "].")

						// config: try to load command configuration
						commandConfig, err := config.GetCommandConfiguration(cmd, filesystem.GetWorkingDirectory(), c.StringSlice("config-include"))
						common.CheckForError(err)

						// container
						containerRuntime := &container_runtime.ContainerRuntime{}
						container := containerRuntime.NewContainer()
						container.SetImage(commandConfig.Image)
						container.PullImage()
					}

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
					log.Debug().Msg("Installing aliases ...")
					scopeFilter := c.String("scope")

					// create global-scoped aliases
					if scopeFilter == "all" || scopeFilter == "global" {
						var globalConfigPath = collection.MapGetValueOrDefault(propConfig.Properties, "global-configuration-path", defaultConfigurationDirectory)
						log.Debug().Msg("Will load the global configuration from [" + globalConfigPath + "].")
						globalConfig, _ := config.LoadProjectConfig(globalConfigPath + "/.envcli.yml")

						for _, element := range globalConfig.Images {
							element.Scope = "Global"
							log.Debug().Msg("Created aliases for " + element.Name + " [Scope: " + element.Scope + "]")

							// for each provided command
							for _, currentCommand := range element.Provides {
								aliases.InstallAlias(currentCommand, element.Scope)
							}
						}
					}

					// create project-scoped aliases
					if scopeFilter == "all" || scopeFilter == "project" {
						var projectDirectory, projectDirectoryErr = config.GetProjectDirectory()
						if projectDirectoryErr != nil && scopeFilter == "project" {
							log.Error().Msg("Can't install project-specific aliases as no valid project was found!")
							os.Exit(1)
						} else if projectDirectoryErr != nil {
							log.Warn().Msg("Can't find a project directory, not throwing a error since all aliases are supposed to be installed!")
						} else {
							log.Debug().Msg("Project Directory: " + projectDirectory)
							projectConfig, _ := config.LoadProjectConfig(projectDirectory + "/.envcli.yml")

							for _, element := range projectConfig.Images {
								element.Scope = "Project"
								log.Debug().Msg("Created aliases for " + element.Name + " [Scope: " + element.Scope + "]")

								// for each provided command
								for _, currentCommand := range element.Provides {
									aliases.InstallAlias(currentCommand, element.Scope)
								}
							}
						}
					}

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
								fmt.Printf("Please provide the variable name and the value you want to set in this format. [envcli config set variable value]\n")
								os.Exit(1)
							}
							varName := c.Args().Get(0)
							varValue := c.Args().Get(1)

							// Set value
							config.SetPropertyConfigEntry(varName, varValue)
							fmt.Printf("Set value of %s to [%s]\n", varName, varValue)

							return nil
						},
					},
					&cli.Command{
						Name: "get",
						Action: func(c *cli.Context) error {
							// Check Parameters
							if c.NArg() != 1 {
								fmt.Printf("Please provide the variable name you want to read. [envcli config get variable]\n")
								os.Exit(1)
							}
							varName := c.Args().Get(0)

							// Get Value
							fmt.Printf("%s [%s]\n", varName, config.GetPropertyConfigEntry(varName))

							return nil
						},
					},
					&cli.Command{
						Name: "get-all",
						Action: func(c *cli.Context) error {
							// Print all values
							for key, value := range propConfig.Properties {
								fmt.Printf("%s [%s]\n", key, value)
							}

							return nil
						},
					},
					&cli.Command{
						Name: "unset",
						Action: func(c *cli.Context) error {
							// Check Parameters
							if c.NArg() != 1 {
								fmt.Printf("Please provide the variable name you want to unset. [envcli config unset variable]\n")
								os.Exit(1)
							}
							varName := c.Args().Get(0)

							// Unset value
							config.UnsetPropertyConfigEntry(varName)
							fmt.Printf("Value of variable %s set to [].\n", varName)

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
