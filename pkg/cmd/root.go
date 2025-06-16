package cmd

import (
	"os"
	"strings"

	"github.com/EnvCLI/EnvCLI/pkg/config"
	"github.com/cidverse/cidverseutils/pkg/collection"
	"github.com/cidverse/cidverseutils/zerologconfig"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	cfg = struct {
		LogLevel  string
		LogFormat string
		LogCaller bool
	}{}
)

var propConfig config.PropertyConfigurationFile

func init() {
	rootCmd.PersistentFlags().StringVar(&cfg.LogLevel, "log-level", "info", "log level - allowed: "+strings.Join(zerologconfig.ValidLogLevels, ","))
	rootCmd.PersistentFlags().StringVar(&cfg.LogFormat, "log-format", "color", "log format - allowed: "+strings.Join(zerologconfig.ValidLogFormats, ","))
	rootCmd.PersistentFlags().BoolVar(&cfg.LogCaller, "log-caller", false, "include caller in log functions")
	rootCmd.PersistentFlags().StringArray("config-include", []string{}, "Additionally include these configuration files, please take note that precedence will be in this order: project config, included, system config")
}

var rootCmd = &cobra.Command{
	Use:   `envcli`,
	Short: "Runs cli commands within docker containers to provide a modern development environment",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// logger
		zerologconfig.Configure(cfg)

		// detect debug mode
		debugValue, debugIsSet := os.LookupEnv("ENVCLI_DEBUG")
		if debugIsSet && strings.ToLower(debugValue) == "true" {
			zerolog.SetGlobalLevel(zerolog.TraceLevel)
		}

		// Global Configuration
		propConfig, propConfigErr := config.LoadPropertyConfig()

		// Configure Proxy Server
		if propConfigErr == nil {
			// Set Proxy Server
			os.Setenv("HTTP_PROXY", collection.MapGetValueOrDefault(propConfig.Properties, "http-proxy", ""))
			os.Setenv("HTTPS_PROXY", collection.MapGetValueOrDefault(propConfig.Properties, "https-proxy", ""))
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(0)
	},
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
