package main

import (
	"os"

	"github.com/EnvCLI/EnvCLI/pkg/cmd"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	status  = "clean"
)

// Init Hook
func init() {
	// Set Version Information
	cmd.Version = version
	cmd.CommitHash = commit
	cmd.BuildAt = date
	cmd.RepositoryStatus = status

	// Initialize Global Logger
	colorableOutput := colorable.NewColorableStdout()
	log.Logger = zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: colorableOutput}).With().Timestamp().Logger()

	// Timestamp Format
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Only log the warning severity or above.
	zerolog.SetGlobalLevel(zerolog.WarnLevel)

	// show calling files
	_, showCalls := os.LookupEnv("ENVCLI_SHOW_CALL")
	if showCalls {
		log.Logger = zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: colorableOutput}).With().Timestamp().Caller().Logger()
	}
}

// CLI Main Entrypoint
func main() {
	// run
	cmdErr := cmd.Execute()
	if cmdErr != nil {
		log.Fatal().Err(cmdErr).Msg("cli error")
	}
}
