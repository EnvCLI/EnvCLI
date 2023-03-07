package cmd

import (
	"fmt"
	"strings"

	"github.com/EnvCLI/EnvCLI/pkg/common"
	"github.com/EnvCLI/EnvCLI/pkg/config"
	"github.com/cidverse/cidverseutils/pkg/containerruntime"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pullImageCmd)
}

var pullImageCmd = &cobra.Command{
	Use:     "pull-image",
	Short:   "pulls the needed images for the specified commands",
	Aliases: []string{},
	Run: func(cmd *cobra.Command, args []string) {
		configIncludes, _ := cmd.PersistentFlags().GetStringArray("config-include")
		fmt.Printf("Pulling images for [%s].\n", strings.Join(args, ", "))

		for _, cmd := range args {
			log.Debug().Msg("Pulling image for command [" + cmd + "].")

			// config: try to load command configuration
			commandConfig, err := config.GetCommandConfiguration(cmd, filesystem.GetWorkingDirectory(), configIncludes)
			common.CheckForError(err)

			// container
			containerRuntime := &containerruntime.ContainerRuntime{}
			container := containerRuntime.NewContainer()
			container.SetImage(commandConfig.Image)
			container.PullImage()
		}
	},
}
