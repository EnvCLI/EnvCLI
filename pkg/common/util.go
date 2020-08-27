package common

import (
	"bytes"
	"runtime"
	"strings"
	"os"

	log "github.com/sirupsen/logrus"
)

/**
 * Sets the loglevel according to the flag on each command run
 */
func SetLoglevel(loglevel string) {
	if loglevel == "panic" {
		log.SetLevel(log.PanicLevel)
		log.SetReportCaller(false)
	} else if loglevel == "fatal" {
		log.SetLevel(log.FatalLevel)
		log.SetReportCaller(false)
	} else if loglevel == "warn" {
		log.SetLevel(log.WarnLevel)
		log.SetReportCaller(false)
	} else if loglevel == "info" {
		log.SetLevel(log.InfoLevel)
		log.SetReportCaller(false)
	} else if loglevel == "debug" {
		log.SetLevel(log.DebugLevel)
		log.SetReportCaller(true)
	} else if loglevel == "trace" {
		log.SetLevel(log.TraceLevel)
		log.SetReportCaller(true)
	}
}

// ParseAndEscapeArgs takes all cli arguments, quotes them and handles escaping
func ParseAndEscapeArgs(args []string) string {
	var commandWithArguments bytes.Buffer
	for _, arg := range args {
		log.Trace("Parsing arg: " + arg)
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

// CheckForError checks if a error happend and logs it, and ends the process
func CheckForError(err error) {
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
