package cmd

import (
	"strconv"
	"time"

	"github.com/EnvCLI/EnvCLI/pkg/updater"
	"github.com/cidverse/cidverseutils/pkg/cihelper"
	"github.com/cidverse/cidverseutils/pkg/collection"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolP("force", "f", false, "A forced update would also redownload the current version.")
	updateCmd.Flags().String("target", "latest", "A target version that should be upgraded/downgraded to.")
}

var updateCmd = &cobra.Command{
	Use:     "self-update",
	Aliases: []string{},
	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		force, _ := cmd.Flags().GetBool("force")

		// Update Check, once a day (not in CI)
		appUpdater := updater.ApplicationUpdater{GitHubOrg: "EnvCLI", GitHubRepository: "EnvCLI"}
		var lastUpdateCheck, _ = strconv.ParseInt(collection.MapGetValueOrDefault(propConfig.Properties, "last-update-check", strconv.Itoa(int(time.Now().Unix()))), 10, 64)
		if time.Now().Unix() >= lastUpdateCheck+86400 && cihelper.IsCIEnvironment() == false {
			if appUpdater.IsUpdateAvailable(cmd.Version) {
				log.Warn().Msg("You are using a old version, please consider to update using `envcli self-update`!")
			}
		}

		appUpdater.Update(target, force, cmd.Version)
	},
}
