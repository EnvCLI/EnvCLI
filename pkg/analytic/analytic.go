package analytic

import (
	sentry "github.com/EnvCLI/EnvCLI/pkg/sentry"
	ogle "github.com/jpillora/go-ogle-analytics"
)

var analticsClient, analticsClientErr = ogle.NewClient("UA-135644097-1")

// Run docker instance
func init() {
	if analticsClientErr != nil {
		// pass error to error handling and ignore, even if metrics don't work it doesn't matter for the user
		sentry.HandleError(analticsClientErr)
	}
}

func TriggerEvent(eventCategory string, eventName string) {
	var err = analticsClient.Send(ogle.NewEvent(eventCategory, eventName))
	if err != nil {
		// pass error to error handling and ignore, even if metrics don't work it doesn't matter for the user
		sentry.HandleError(err)
	}
}
