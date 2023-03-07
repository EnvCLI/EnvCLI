package cmd

import (
	"os"
	"strings"

	"github.com/EnvCLI/EnvCLI/pkg/config"
	"github.com/cidverse/cidverseutils/pkg/collection"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/thoas/go-funk"
)

var (
	cfg = struct {
		LogLevel  string
		LogFormat string
		LogCaller bool
	}{}
	validLogLevels  = []string{"trace", "debug", "info", "warn", "error"}
	validLogFormats = []string{"plain", "color", "json"}
)

var propConfig config.PropertyConfigurationFile

func init() {
	rootCmd.PersistentFlags().StringVar(&cfg.LogLevel, "log-level", "info", "log level - allowed: "+strings.Join(validLogLevels, ","))
	rootCmd.PersistentFlags().StringVar(&cfg.LogFormat, "log-format", "color", "log format - allowed: "+strings.Join(validLogFormats, ","))
	rootCmd.PersistentFlags().BoolVar(&cfg.LogCaller, "log-caller", false, "include caller in log functions")
	rootCmd.PersistentFlags().StringArray("config-include", []string{}, "Additionally include these configuration files, please take note that precedence will be in this order: project config, included, system config")
}

var rootCmd = &cobra.Command{
	Use:   `envcli`,
	Short: "Runs cli commands within docker containers to provide a modern development environment",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// log format
		if !funk.ContainsString(validLogFormats, cfg.LogFormat) {
			log.Error().Str("current", cfg.LogFormat).Strs("valid", validLogFormats).Msg("invalid log format specified")
			os.Exit(1)
		}
		var logContext zerolog.Context
		if cfg.LogFormat == "plain" {
			logContext = zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true}).With().Timestamp()
		} else if cfg.LogFormat == "color" {
			colorableOutput := colorable.NewColorableStdout()
			logContext = zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: colorableOutput, NoColor: false}).With().Timestamp()
		} else if cfg.LogFormat == "json" {
			logContext = zerolog.New(os.Stderr).Output(os.Stderr).With().Timestamp()
		}
		if cfg.LogCaller {
			logContext = logContext.Caller()
		}
		log.Logger = logContext.Logger()

		// log time format
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

		// detect debug mode
		debugValue, debugIsSet := os.LookupEnv("ENVCLI_DEBUG")
		if debugIsSet && strings.ToLower(debugValue) == "true" {
			zerolog.SetGlobalLevel(zerolog.TraceLevel)
		}

		// log level
		if !funk.ContainsString(validLogLevels, cfg.LogLevel) {
			log.Error().Str("current", cfg.LogLevel).Strs("valid", validLogLevels).Msg("invalid log level specified")
			os.Exit(1)
		}
		if cfg.LogLevel == "trace" {
			zerolog.SetGlobalLevel(zerolog.TraceLevel)
		} else if cfg.LogLevel == "debug" {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		} else if cfg.LogLevel == "info" {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		} else if cfg.LogLevel == "warn" {
			zerolog.SetGlobalLevel(zerolog.WarnLevel)
		} else if cfg.LogLevel == "error" {
			zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		}

		// logging config
		log.Debug().Str("log-level", cfg.LogLevel).Str("log-format", cfg.LogFormat).Bool("log-caller", cfg.LogCaller).Msg("configured logging")

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
