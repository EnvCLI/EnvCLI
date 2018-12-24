package main

import (
	"os"
	"runtime"

	log "github.com/sirupsen/logrus" // imports as package "log"
)

/**
 * CLI Command Passthru with input/output
 */
func installAlias(command string, scope string) error {
	log.Debugf("Creating alias for command: %s [Scope: %s]", command, scope)

	// download alias script for each used command
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		log.Debugf("Detected Linux - Will place bash scripts into PATH ...")
		aliasScriptURL := "https://raw.githubusercontent.com/EnvCLI/EnvCLI/" + appVersion + "/scripts/alias.sh"
		aliasScriptFilepath := configurationLoader.GetExecutionDirectory() + "/" + command

		err := DownloadFile(aliasScriptFilepath, aliasScriptURL)
		if err != nil {
			log.Errorf("Can't create alias [%s], download failed.", command)
			panic(err)
		} else {
			log.Debugf("Created alias for [%s]!", command)

			// set execute permissions
			chmodErr := os.Chmod(aliasScriptFilepath, 0744)
			if chmodErr != nil {
				log.Errorf("Failed to make the alias script for [%s] executable!", command)
			} else {
				log.Debugf("Made alias script for [%s] executable!", command)
			}
		}
	} else if runtime.GOOS == "windows" {
		log.Debugf("Detected Windows - Will place cmd scripts into PATH ...")
		aliasScriptURL := "https://raw.githubusercontent.com/EnvCLI/EnvCLI/" + appVersion + "/scripts/alias.cmd"
		aliasScriptFilepath := configurationLoader.GetExecutionDirectory() + "/" + command + ".cmd"

		err := DownloadFile(aliasScriptFilepath, aliasScriptURL)
		if err != nil {
			log.Errorf("Can't create alias [%s], download failed.", command)
			panic(err)
		} else {
			log.Debugf("Created alias for [%s]!", command)
		}
	} else {
		log.Errorf("Can't create alias for [%s]. Aliases aren't supported on %s yet!", command, runtime.GOOS)
	}

	return nil
}
