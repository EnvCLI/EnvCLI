package aliases

import (
	"os"
	"runtime"

	sentry "github.com/EnvCLI/EnvCLI/pkg/sentry"
	util "github.com/EnvCLI/EnvCLI/pkg/util"
	log "github.com/sirupsen/logrus" // imports as package "log"
)

/**
 * CLI Command Passthru with input/output
 */
func InstallAlias(appVersion string, command string, scope string) error {
	log.Debugf("Creating alias for command: %s [Scope: %s]", command, scope)

	// download alias script for each used command
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		log.Debugf("Detected Linux - Will place bash scripts into PATH ...")
		aliasScriptURL := "https://raw.githubusercontent.com/EnvCLI/EnvCLI/" + appVersion + "/scripts/alias.sh"
		aliasScriptFilepath := util.GetExecutionDirectory() + "/" + command

		err := util.DownloadFile(aliasScriptFilepath, aliasScriptURL)
		if err != nil {
			log.Errorf("Can't create alias [%s], download failed.", command)
			sentry.HandleError(err)
			panic(err)
		} else {
			log.Debugf("Created alias for [%s]!", command)

			// set execute permissions
			chmodErr := os.Chmod(aliasScriptFilepath, 0744)
			if chmodErr != nil {
				log.Errorf("Failed to make the alias script for [%s] executable!", command)
				sentry.HandleError(chmodErr)
			} else {
				log.Debugf("Made alias script for [%s] executable!", command)
			}
		}
	} else if runtime.GOOS == "windows" {
		log.Debugf("Detected Windows - Will place cmd scripts into PATH ...")
		aliasScriptURL := "https://raw.githubusercontent.com/EnvCLI/EnvCLI/" + appVersion + "/scripts/alias.cmd"
		aliasScriptFilepath := util.GetExecutionDirectory() + "/" + command + ".cmd"

		err := util.DownloadFile(aliasScriptFilepath, aliasScriptURL)
		if err != nil {
			log.Errorf("Can't create alias [%s], download failed.", command)
			sentry.HandleError(err)
			panic(err)
		} else {
			log.Debugf("Created alias for [%s]!", command)
		}
	} else {
		log.Errorf("Can't create alias for [%s]. Aliases aren't supported on %s yet!", command, runtime.GOOS)
	}

	return nil
}
