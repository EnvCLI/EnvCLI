package cmd

import (
	"os"

	"github.com/EnvCLI/EnvCLI/pkg/aliases"
	"github.com/EnvCLI/EnvCLI/pkg/config"
	"github.com/cidverse/cidverseutils/pkg/collection"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(installAliasesCmd)
	installAliasesCmd.Flags().StringP("scope", "s", "all", "Install aliases for the specified scope (project, global or all)")
}

var installAliasesCmd = &cobra.Command{
	Use:     "install-aliases",
	Short:   "installs aliases for the global / project scoped commands",
	Aliases: []string{},
	Run: func(cmd *cobra.Command, args []string) {
		scopeFilter, _ := cmd.Flags().GetString("scope")
		log.Debug().Msg("Installing aliases ...")

		// create global-scoped aliases
		if scopeFilter == "all" || scopeFilter == "global" {
			var globalConfigPath = collection.MapGetValueOrDefault(propConfig.Properties, "global-configuration-path", filesystem.GetExecutionDirectory())
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
	},
}
