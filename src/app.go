package main

import (
	"os"
	"time"
	"strings"
	"sort"
	"runtime"
	log "github.com/sirupsen/logrus" // imports as package "log"
	"gopkg.in/urfave/cli.v2" // imports as package "cli"
	"github.com/mattn/go-colorable" // imports as package "colorable"
)

var appName string = "EnvCLI Utility"
var appVersion string = "v0.1.0"

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
	configurationLoader := ConfigurationLoader{}
	globalConfig, err := configurationLoader.loadGlobalConfig(configurationLoader.getExecutionDirectory() + "/.envclirc")

	// Configure Proxy Server
	if(err == nil) {
		// Set Proxy Server
		os.Setenv("HTTP_PROXY", globalConfig.HttpProxy)
		os.Setenv("HTTPS_PROXY", globalConfig.HttpsProxy)
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
			&cli.StringSliceFlag{
				Name: "env",
				Aliases: []string{"e"},
				Usage: "Sets environment variables within the containers",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "self-update",
				Aliases: []string{},
				Usage:   "updates the dev cli utility",
				Flags: []cli.Flag {
		      &cli.BoolFlag{
		        Name: "force",
						Aliases: []string{"f"},
		        Value: false,
		        Usage: "A forced update would also redownload the current version",
		      },
		    },
				Action: func(c *cli.Context) error {
					// Set loglevel
					setLoglevel(c.String("loglevel"))

					// Run Update
					appUpdater := ApplicationUpdater{BintrayOrg: "envcli", BintrayRepository: "golang", BintrayPackage: "envcli", GitHubOrg: "PhilippHeuer", GitHubRepository: "EnvCLI"}
					appUpdater.update("latest", c.Bool("force"))

					return nil
				},
			},
			{
				Name:    "run",
				Aliases: []string{},
				Usage:   "runs 3rd party commands within their respective docker containers",
				Action: func(c *cli.Context) error {
					// Set loglevel
					setLoglevel(c.String("loglevel"))

					// parse command
					commandName := c.Args().First()
					commandWithArguments := strings.Join(append([]string{commandName}, c.Args().Tail()...), " ")
					log.Debugf("Recieved request to run command [%s] - with Arguments [%s].", commandName, commandWithArguments)

					// load yml project configuration
					configurationLoader := ConfigurationLoader{}
					if configurationLoader.getProjectDirectory() == "" {
						log.Fatalf("No .envcli.yml configration file found in current or parent directories. Please run envcli within your project.")
						return nil
					}
					config, _ := configurationLoader.loadProjectConfig(configurationLoader.getProjectDirectory() + "/.envcli.yml")

					// check for command prefix and get the matching configuration entry
					var dockerImage string = ""
					var dockerImageTag string = ""
					var projectDirectory string
					var commandShell string = ""
					for _, element := range config.Commands {
						log.Debugf("Checking for matching commands in package %s", element.Name)
						for _, providedCommand := range element.Provides {
							log.Debugf("Comparing used command [%s] with provided command %s of %s", commandName, providedCommand, element.Name)
							if providedCommand == commandName {
								log.Debugf("Matched command %s against package [%s]", commandName, element.Name)
								dockerImage = element.Image
								dockerImageTag = element.Tag
								projectDirectory = element.Directory
								commandShell = element.Shell
								log.Debugf("Image: %s | Tag: %s | ImageDirectory: %s", dockerImage, dockerImageTag, projectDirectory)
							}
						}
					}
					if dockerImage == "" {
						log.Errorf("No configuration for command [%s] found.", commandName)
						return nil
					}

					// detect container service and send command
					log.Infof("Executing specified command in Docker Container [%s:%s].", dockerImage, dockerImageTag)
					docker := Docker{}
					docker.containerExec(dockerImage, dockerImageTag, commandShell, commandWithArguments, configurationLoader.getProjectDirectory(), projectDirectory, projectDirectory+"/"+configurationLoader.getRelativePathToWorkingDirectory(), c.StringSlice("env"))

					return nil
				},
			},
			{
				Name:    "config",
				Aliases: []string{},
				Usage:   "updates the dev cli utility",
				Subcommands: []*cli.Command{
				  &cli.Command{
						Name:   "set",
						Action: func(c *cli.Context) error {
							// Set loglevel
							setLoglevel(c.String("loglevel"))

							// Load Config
							configurationLoader := ConfigurationLoader{}
							globalConfig, _ := configurationLoader.loadGlobalConfig(configurationLoader.getExecutionDirectory() + "/.envclirc")

							// Check Parameters
							if c.NArg() != 2 {
					      log.Fatal("Please provide the variable name and the value you want to set in this format. [envcli config set variable value]")
							}
							varName := c.Args().Get(0)
							varValue := c.Args().Get(1)

							// Set Value
							if varName == "http-proxy" {
								globalConfig.HttpProxy = varValue
								log.Infof("Set value of HttpProxy to [%s]", globalConfig.HttpProxy)
							} else if varName == "https-proxy" {
								globalConfig.HttpsProxy = varValue
								log.Infof("Set value of HttpsProxy to [%s]", globalConfig.HttpsProxy)
							} else {
								log.Infof("Unknown variable name [%s]", varName)
							}

							// Save Config
							configurationLoader.saveGlobalConfig(configurationLoader.getExecutionDirectory() + "/.envclirc", globalConfig)

							return nil
						},
				  },
					&cli.Command{
						Name:   "get",
						Action: func(c *cli.Context) error {
							// Set loglevel
							setLoglevel(c.String("loglevel"))

							// Load Config
							configurationLoader := ConfigurationLoader{}
							globalConfig, _ := configurationLoader.loadGlobalConfig(configurationLoader.getExecutionDirectory() + "/.envclirc")

							// Check Parameters
							if c.NArg() != 1 {
					      log.Fatal("Please provide the variable name you want to erase. [envcli config unset variable]")
							}
							varName := c.Args().Get(0)

							// Get Value
							if varName == "http-proxy" {
								log.Infof("HttpProxy [%s]", globalConfig.HttpProxy)
							} else if varName == "https-proxy" {
								log.Infof("HttpsProxy [%s]", globalConfig.HttpsProxy)
							} else {
								log.Infof("Unknown variable name [%s]", varName)
							}

							return nil
						},
				  },
					&cli.Command{
						Name:   "unset",
						Action: func(c *cli.Context) error {
							// Set loglevel
							setLoglevel(c.String("loglevel"))

							// Load Config
							configurationLoader := ConfigurationLoader{}
							globalConfig, _ := configurationLoader.loadGlobalConfig(configurationLoader.getExecutionDirectory() + "/.envclirc")

							// Check Parameters
							if c.NArg() != 1 {
					      log.Fatal("Please provide the variable name you want to read. [envcli config get variable]")
							}
							varName := c.Args().Get(0)

							// Get Value
							if varName == "http-proxy" {
								globalConfig.HttpProxy = ""
								log.Info("Unset variable HttpProxy.")
							} else if varName == "https-proxy" {
								globalConfig.HttpsProxy = ""
								log.Info("Unset variable HttpsProxy.")
							} else {
								log.Infof("Unknown variable name [%s]", varName)
							}

							// Save Config
							configurationLoader.saveGlobalConfig(configurationLoader.getExecutionDirectory() + "/.envclirc", globalConfig)

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
