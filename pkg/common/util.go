package common

import (
	"bytes"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
)

/**
 * Sets the loglevel according to the flag on each command run
 */
func SetLoglevel(loglevel string) {
	if loglevel == "panic" {
		log.SetLevel(log.PanicLevel)
	} else if loglevel == "fatal" {
		log.SetLevel(log.FatalLevel)
	} else if loglevel == "warn" {
		log.SetLevel(log.WarnLevel)
	} else if loglevel == "info" {
		log.SetLevel(log.InfoLevel)
	} else if loglevel == "debug" {
		log.SetLevel(log.DebugLevel)
	} else if loglevel == "trace" {
		log.SetLevel(log.TraceLevel)
	}
}

// ParseAndEscapeArgs takes all cli arguments, quotes them and handles escaping
func ParseAndEscapeArgs(args []string) string {
	var commandWithArguments bytes.Buffer
	for _, arg := range args {
		log.Debug("Parsing arg: " + arg)
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
