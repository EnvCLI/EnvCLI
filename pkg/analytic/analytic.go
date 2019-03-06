package analytic

import (
	sentry "github.com/EnvCLI/EnvCLI/pkg/sentry"
	machineid "github.com/denisbrodbeck/machineid"
	ogle "github.com/jpillora/go-ogle-analytics"
)

var analticsClient, analticsClientErr = ogle.NewClient("UA-135644097-1")

/**
 * Initialize
 * Docs: https://godoc.org/github.com/jpillora/go-ogle-analytics
 */
func InitializeAnalytics(appName string, appVersion string) {
	// check if we initialized the analticsClient successfully
	if analticsClientErr != nil {
		// pass error to error handling and ignore, even if metrics don't work it doesn't matter for the user
		sentry.HandleError(analticsClientErr)
	} else {
		// App Info
		analticsClient = analticsClient.ApplicationName(appName)
		analticsClient = analticsClient.ApplicationVersion(appVersion)

		// Unique User Id
		deviceid, deviceidErr := machineid.ProtectedID(appName)
		if deviceidErr != nil {
			sentry.HandleError(deviceidErr)
		} else {
			analticsClient = analticsClient.UserID(deviceid)
		}

		// Locale
		var systemLocale, systemLocaleErr = GetSystemLocale()
		if systemLocaleErr != nil {
			// pass error to error handling and ignore, even if metrics don't work it doesn't matter for the user
			sentry.HandleError(analticsClientErr)
		} else {
			analticsClient = analticsClient.UserLanguage(systemLocale)
		}
	}
}

func TriggerEvent(eventCategory string, eventName string) {
	var err = analticsClient.Send(ogle.NewEvent(eventCategory, eventName))
	if err != nil {
		// pass error to error handling and ignore, even if metrics don't work it doesn't matter for the user
		sentry.HandleError(err)
	}
}
