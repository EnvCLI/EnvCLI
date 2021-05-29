package aliases

import (
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"runtime"

	common "github.com/EnvCLI/EnvCLI/pkg/common"
)

// InstallAlias installs simple aliases that pass all parameters to envcli run
func InstallAlias(command string, scope string) error {
	log.Debug().Str("command", command).Str("scope", scope).Msg("Installing alias ...")

	// download alias script for each used command
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		log.Debug().Msg("Detected Linux - Will place bash scripts into PATH ...")

		scriptData, err := Asset("scripts/alias.sh")
		common.CheckForError(err)

		err = ioutil.WriteFile(filesystem.GetExecutionDirectory()+"/"+command, scriptData, 0755)
		common.CheckForError(err)

		log.Debug().Str("command", command).Msg("Installed alias!")
	} else if runtime.GOOS == "windows" {
		log.Debug().Msg("Detected Windows - Will place cmd scripts into PATH ...")

		scriptData, err := Asset("scripts/alias.cmd")
		common.CheckForError(err)

		err = ioutil.WriteFile(filesystem.GetExecutionDirectory()+"/"+command+".cmd", scriptData, 0755)
		common.CheckForError(err)

		log.Debug().Str("command", command).Msg("Installed alias!")
	} else {
		log.Error().Str("command", command).Str("platform",  runtime.GOOS).Msg("Failed to install alias! Not supported on current platform.")
	}

	return nil
}
