package main

import (
	"os"
	"os/exec"
	"time"
	"fmt"
	"strings"
	"sort"
	"runtime"
	log "github.com/sirupsen/logrus" // imports as package "log"
	"gopkg.in/urfave/cli.v2" // imports as package "cli"
)

// Init Hook
func init() {
	log.SetOutput(os.Stdout)
  log.SetLevel(log.DebugLevel)
}

// CLI Main Entrypoint
func main() {
	app := &cli.App{
    Name: "EnvCLI Utility",
		Version: "v0.1.1",
    Compiled: time.Now(),
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
        Name:    "run",
        Aliases: []string{},
        Usage:   "runs 3rd party commands within their respective docker containers",
        Action:  func(c *cli.Context) error {
					// Set loglevel
					setLoglevel(c.String("loglevel"))

					// parse command
					commandName := c.Args().First()
					commandWithArguments := strings.Join(append([]string{commandName}, c.Args().Tail()...), " ")
					log.Debugf("Command run in Remote: %s | %s", commandName, commandWithArguments)

					// load yml project configuration
					configurationLoader := ConfigurationLoader{}
					var config ProjectConfigrationFile = configurationLoader.load(configurationLoader.getWorkingDirectory() + "/.envcli.yml")

					// check for command prefix and get the matching configuration entry
					var dockerImage string = ""
					var dockerImageTag string = ""
					var projectDirectory string
					for _, element := range config.Commands {
						if element.Name == commandName {
							log.Debugf("Found matching entry in configuration for command %s [%s]", commandName, element.Description)
							dockerImage = element.Image
							dockerImageTag = element.Tag
							projectDirectory = element.ProjectDirectory
							log.Debugf("Image: %s | Tag: %s", dockerImage, dockerImageTag)
						}
					}
					if dockerImage == "" {
						log.Debugf("No configuration for command [%s] found.", commandName)
						return nil
					}

					// detect container service and send command
					// - docker for windows
					if runtime.GOOS == "windows" {
						log.Infof("Redirecting command to Docker Container [%s:%s].", dockerImage, dockerImageTag)
						var dockerCommand string = fmt.Sprintf("docker run --rm --volume \"%s:%s\" -w /project %s:%s %s", configurationLoader.getWorkingDirectory(), projectDirectory, dockerImage, dockerImageTag, commandWithArguments)
						log.Debugf("Docker Command: %s", dockerCommand)
						execCommandWithResponse(dockerCommand)
					}

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

func execCommandWithResponse(command string) {
	var commandPrefix string
	if runtime.GOOS == "windows" {
		commandPrefix = "powershell"
	} else {
		commandPrefix = ""
	}

	cmd := exec.Command(commandPrefix, command)
	cmd.Stdout = os.Stdout
	if err := cmd.Start(); err != nil {
			log.Fatal(err)
	}
	cmd.Wait()
}
