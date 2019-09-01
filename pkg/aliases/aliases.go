package aliases

import (
	"io/ioutil"
	"runtime"

	common "github.com/EnvCLI/EnvCLI/pkg/common"
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

		scriptData, err := Asset("scripts/alias.sh")
		common.CheckForError(err)

		err = ioutil.WriteFile(util.GetExecutionDirectory()+"/"+command, scriptData, 0755)
		common.CheckForError(err)

		log.Debugf("Created alias for [%s]!", command)
	} else if runtime.GOOS == "windows" {
		log.Debugf("Detected Windows - Will place cmd scripts into PATH ...")

		scriptData, err := Asset("scripts/alias.cmd")
		common.CheckForError(err)

		err = ioutil.WriteFile(util.GetExecutionDirectory()+"/"+command+".cmd", scriptData, 0755)
		common.CheckForError(err)

		log.Debugf("Created alias for [%s]!", command)
	} else {
		log.Errorf("Can't create alias for [%s]. Aliases aren't supported on %s yet!", command, runtime.GOOS)
	}

	return nil
}
