package common

import (
	"bytes"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"runtime"
	"strings"
)

/**
 * Sets the loglevel according to the flag on each command run
 */
func SetLoglevel(loglevel string) {
	if loglevel == "panic" {
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	} else if loglevel == "fatal" {
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	} else if loglevel == "warn" {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	} else if loglevel == "info" {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else if loglevel == "debug" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else if loglevel == "trace" {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}
}

// ParseAndEscapeArgs takes all cli arguments, quotes them and handles escaping
func ParseAndEscapeArgs(args []string) string {
	var commandWithArguments bytes.Buffer
	for _, arg := range args {
		log.Trace().Msg("Parsing arg: " + arg)
		if runtime.GOOS == "windows" {
			quotedArg := "\"" + strings.Replace(strings.Trim(arg, "\""), "\"", "`\"", -1) + "\""
			commandWithArguments.WriteString(quotedArg + " ")
		} else {
			quotedArg := "\"" + strings.Replace(strings.Trim(arg, "\""), "\"", "\\\"", -1) + "\""
			commandWithArguments.WriteString(quotedArg + " ")
		}
	}

	command := commandWithArguments.String()
	return command[:len(command)-1]
}

// CheckForError checks if a error happened and logs it, and ends the process
func CheckForError(err error) {
	if err != nil {
		log.Fatal().Err(err).Msg(err.Error())
	}
}
