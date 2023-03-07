package cmd

import (
	"strings"

	"github.com/EnvCLI/EnvCLI/pkg/common"
	"github.com/EnvCLI/EnvCLI/pkg/config"
	"github.com/cidverse/cidverseutils/pkg/cihelper"
	"github.com/cidverse/cidverseutils/pkg/collection"
	"github.com/cidverse/cidverseutils/pkg/containerruntime"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
	installAliasesCmd.Flags().StringArrayP("env", "e", []string{}, "Sets environment variables within the containers")
	installAliasesCmd.Flags().StringArrayP("port", "p", []string{}, "Publish ports of the container")
	installAliasesCmd.Flags().StringArray("userArgs", []string{}, "Allows to specify custom arguments that will be passed to the docker run command for special cases")
}

var runCmd = &cobra.Command{
	Use:     "run",
	Short:   "runs 3rd party commands within their respective docker containers",
	Aliases: []string{},
	Run: func(cmd *cobra.Command, args []string) {
		env, _ := cmd.Flags().GetStringArray("env")
		port, _ := cmd.Flags().GetStringArray("port")
		userArgs, _ := cmd.Flags().GetStringArray("userArgs")
		configIncludes, _ := cmd.PersistentFlags().GetStringArray("config-include")

		// parse command
		commandName := args[0]

		// iterate and quote args if needed
		commandArgs := append([]string{commandName}, strings.Join(args[1:], " "))
		commandWithArguments := common.ParseAndEscapeArgs(commandArgs)

		log.Debug().Msg("Received request to run command [" + commandName + "] - with Arguments [" + commandWithArguments + "].")

		// config: try to load command configuration
		commandConfig, commandConfigErr := config.GetCommandConfiguration(commandName, filesystem.GetWorkingDirectory(), configIncludes)
		if commandConfigErr != nil {
			log.Fatal().Err(commandConfigErr).Msg("failed to load command config")
		}

		// container runtime
		containerRuntime := &containerruntime.ContainerRuntime{}
		container := containerRuntime.NewContainer()
		container.SetImage(commandConfig.Image)
		container.SetEntrypoint(commandConfig.Entrypoint)
		container.SetCommandShell(commandConfig.Shell)

		// mounts
		projectOrExecutionDir := config.GetProjectOrWorkingDirectory()
		mountDir := commandConfig.Directory
		if mountDir == "" {
			mountDir = projectOrExecutionDir
		}
		log.Debug().Str("source", projectOrExecutionDir).Str("target", mountDir).Msg("Adding volume mount")
		container.AddVolume(containerruntime.ContainerMount{MountType: "directory", Source: projectOrExecutionDir, Target: mountDir})
		container.SetWorkingDirectory(commandConfig.Directory + "/" + filesystem.GetPathRelativeToDirectory(filesystem.GetWorkingDirectory(), projectOrExecutionDir))

		// core: expose ports (command args)
		container.AddContainerPorts(port)

		// core: pass environment variables (command args)
		container.AddEnvironmentVariables(env)

		// feature: user args
		if len(userArgs) > 0 {
			container.SetUserArgs(strings.Join(userArgs, " "))
		}

		// feature: before_script
		var commandWithBeforeScript = ""
		commandWithBeforeScript = strings.TrimSpace(commandWithArguments)
		if commandConfig.BeforeScript != nil {
			commandWithBeforeScript = strings.Join(commandConfig.BeforeScript[:], ";") + " && " + commandWithBeforeScript

			commandWithBeforeScript = strings.Replace(commandWithBeforeScript, "{HTTPProxy}", collection.MapGetValueOrDefault(propConfig.Properties, "http-proxy", ""), -1)
			commandWithBeforeScript = strings.Replace(commandWithBeforeScript, "{HTTPSProxy}", collection.MapGetValueOrDefault(propConfig.Properties, "https-proxy", ""), -1)
		}
		log.Debug().Msg("Setting new command with before_script: " + commandWithBeforeScript)
		container.SetCommand(commandWithBeforeScript)

		// feature: container runtime access
		if commandConfig.ContainerRuntimeAccess {
			container.AllowContainerRuntimeAcccess()
		}

		// feature: caching
		for _, cachingEntry := range commandConfig.Caching {
			if collection.MapGetValueOrDefault(propConfig.Properties, "cache-path", "") == "" {
				log.Warn().Msg("Cache is disabled, CachePath not set.")
				break
			}

			var cacheFolder = collection.MapGetValueOrDefault(propConfig.Properties, "cache-path", "") + "/" + cachingEntry.Name
			filesystem.CreateDirectory(cacheFolder)
			container.AddCacheMount(cachingEntry.Name, cacheFolder, cachingEntry.ContainerDirectory)
		}

		// feature: capabilities
		for _, cap := range commandConfig.CapAdd {
			container.AddCapability(cap)
		}

		// feature: pass all env variables (excludes system variables like PATH, ...) in CI environments
		if cihelper.IsCIEnvironment() {
			container.AddAllEnvironmentVariables()
		}

		// feature: proxy environment
		httpProxy := collection.MapGetValueOrDefault(propConfig.Properties, "http-proxy", "")
		if httpProxy != "" {
			container.AddEnvironmentVariable("http_proxy", httpProxy)
		}

		httpsProxy := collection.MapGetValueOrDefault(propConfig.Properties, "https-proxy", "")
		if httpsProxy != "" {
			container.AddEnvironmentVariable("https_proxy", httpsProxy)
		}

		// detect container service and send command
		log.Info().Msg("Executing command in container [" + commandConfig.Image + "].")
		container.StartContainer()
	},
}
