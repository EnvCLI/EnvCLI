package analytic

import (
	"runtime"

	sentry "github.com/EnvCLI/EnvCLI/pkg/sentry"
	util "github.com/EnvCLI/EnvCLI/pkg/util"
	machineid "github.com/denisbrodbeck/machineid"
	segment "gopkg.in/segmentio/analytics-go.v3"
)

var analyticsEnabled = false

var segmentClient = segment.New("6B5KsCRcmam7tIFqpVqEStod6QAo4Ttp")
var uniqueId = "random"

/**
 * InitializeAnalytics initialized the analytics client
 * Docs: https://godoc.org/github.com/jpillora/go-ogle-analytics
 */
func InitializeAnalytics(appName string, appVersion string) {
	analyticsEnabled = true

	// Unique User Id
	deviceid, deviceidErr := machineid.ProtectedID(appName)
	if deviceidErr != nil {
		sentry.HandleError(deviceidErr)
	} else {
		uniqueId = deviceid
	}

	// Locale
	var systemLocale, systemLocaleErr = GetSystemLocale()
	if systemLocaleErr != nil {
		// pass error to error handling and ignore, even if metrics don't work it doesn't matter for the user
		sentry.HandleError(systemLocaleErr)
		systemLocale = "XX_XX"
	}

	// Platform
	var platform = "Desktop"
	if util.IsCIEnvironment() {
		platform = "CI"
	}

	// Segment Info
	if analyticsEnabled {
		segmentClient.Enqueue(segment.Identify{
			UserId: uniqueId,
			Traits: segment.NewTraits().
				SetName(uniqueId).
				Set("name", appName).
				Set("version", appVersion).
				Set("os", runtime.GOOS).
				Set("platform", platform).
				Set("locale", systemLocale),
		})
	}
}

/**
 * TriggerEvent triggers a tracked Event which will be visible in analytics
 */
func TriggerEvent(eventName string, eventPayload string) {
	// do nothing if opted-out
	if analyticsEnabled {
		// segmentio
		segmentClient.Enqueue(segment.Track{
			UserId: uniqueId,
			Event:  eventName,
			Properties: segment.NewProperties().
				Set("payload", eventPayload),
		})
	}
}
