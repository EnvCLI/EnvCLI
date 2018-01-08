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

// Init Hook
func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	// Fix color output for windows [https://github.com/Sirupsen/logrus/issues/172]
	if runtime.GOOS == "windows" {
		log.SetFormatter(&log.TextFormatter{ForceColors: true})
		log.SetOutput(colorable.NewColorableStdout())
	}
}

// CLI Main Entrypoint
func main() {
	app := &cli.App{
		Name:                  "EnvCLI Utility",
		Version:               "v0.1.2",
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
		Commands: []*cli.Command{
			{
				Name:    "self-update",
				Aliases: []string{},
				Usage:   "updates the dev cli utility",
				Action: func(c *cli.Context) error {
					// Set loglevel
					setLoglevel(c.String("loglevel"))

					// Run Update
					appUpdater := ApplicationUpdater{AppId: "app_8piLcd8unVA", PublicKey: `-----BEGIN ECDSA PUBLIC KEY-----
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEinl1s7+5o65K2NkavhUP97ZInqs228+e
AbS0hVCbHGFpZfjWHH59MCX0ekQnqDpgkJkHGGsT+gPIGGAIzb52K5T8rq2dbrGg
mmYdo1ZNtsh4rk9sJbQb2IkjSm+n+Xwr
-----END ECDSA PUBLIC KEY-----`}
					appUpdater.update()

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
					log.Debugf("Command run in Remote: %s | %s", commandName, commandWithArguments)

					// load yml project configuration
					configurationLoader := ConfigurationLoader{}
					if configurationLoader.getProjectDirectory() == "" {
						log.Fatalf("No .envcli.yml configration file found in current or parent directories. Please run envcli within your project.")
						return nil
					}
					var config ProjectConfigrationFile = configurationLoader.load(configurationLoader.getProjectDirectory() + "/.envcli.yml")

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
						log.Debugf("No configuration for command [%s] found.", commandName)
						return nil
					}
					
					// detect container service and send command
					log.Infof("Redirecting command to Docker Container [%s:%s].", dockerImage, dockerImageTag)
					docker := Docker{}
					// - docker toolbox (docker-machine)
					if docker.isDockerToolbox() {
						docker.containerExec(dockerImage, dockerImageTag, commandShell, commandWithArguments, configurationLoader.getProjectDirectory(), projectDirectory, projectDirectory+"/"+configurationLoader.getRelativePathToWorkingDirectory())
						return nil
					}
					// - docker native (docker for windows/mac/linux)
					if docker.isDockerNative() {
						docker.containerExec(dockerImage, dockerImageTag, commandShell, commandWithArguments, configurationLoader.getProjectDirectory(), projectDirectory, projectDirectory+"/"+configurationLoader.getRelativePathToWorkingDirectory())
						return nil
					}

					log.Fatal("No supported docker installation found.")
					return nil
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

/**
 * Sets the loglevel according to the flag on each command run
 */
func setLoglevel(loglevel string) {
	if loglevel == "info" {
		log.SetLevel(log.InfoLevel)
	} else if loglevel == "debug" {
		log.SetLevel(log.DebugLevel)
	}
}
