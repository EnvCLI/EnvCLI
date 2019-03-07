package analytic

import (
	sentry "github.com/EnvCLI/EnvCLI/pkg/sentry"
	machineid "github.com/denisbrodbeck/machineid"
	ogle "github.com/jpillora/go-ogle-analytics"
)

var analyticsClient, analyticsClientErr = ogle.NewClient("UA-135644097-1")
var analyticsEnabled = false

/**
 * InitializeAnalytics initialized the analytics client
 * Docs: https://godoc.org/github.com/jpillora/go-ogle-analytics
 */
func InitializeAnalytics(appName string, appVersion string) {
	analyticsEnabled = true

	// check if we initialized the analticsClient successfully
	if analyticsClientErr != nil {
		// pass error to error handling and ignore, even if metrics don't work it doesn't matter for the user
		sentry.HandleError(analyticsClientErr)
	} else {
		// App Info
		analyticsClient = analyticsClient.ApplicationName(appName)
		analyticsClient = analyticsClient.ApplicationVersion(appVersion)

		// Unique User Id
		deviceid, deviceidErr := machineid.ProtectedID(appName)
		if deviceidErr != nil {
			sentry.HandleError(deviceidErr)
		} else {
			analyticsClient = analyticsClient.UserID(deviceid)
		}

		// Locale
		var systemLocale, systemLocaleErr = GetSystemLocale()
		if systemLocaleErr != nil {
			// pass error to error handling and ignore, even if metrics don't work it doesn't matter for the user
			sentry.HandleError(systemLocaleErr)
		} else {
			analyticsClient = analyticsClient.UserLanguage(systemLocale)
		}
	}
}

/**
 * TriggerEvent triggers a tracked Event which will be visible in analytics
 */
func TriggerEvent(eventCategory string, eventName string) {
	// do nothing if opted-out
	if analyticsEnabled {
		var err = analyticsClient.Send(ogle.NewEvent(eventCategory, eventName))
		if err != nil {
			// pass error to error handling and ignore, even if metrics don't work it doesn't matter for the user
			sentry.HandleError(err)
		}
	}
}
