package analytic

import (
	"runtime"
	"strconv"

	sentry "github.com/EnvCLI/EnvCLI/pkg/sentry"
	util "github.com/EnvCLI/EnvCLI/pkg/util"
	machineid "github.com/denisbrodbeck/machineid"
	segment "gopkg.in/segmentio/analytics-go.v3"
)

var analyticsEnabled = false
var segmentClient segment.Client
var segmentClientErr error
var uniqueId = GetHostname()

/**
 * InitializeAnalytics initialized the analytics client
 * Docs: https://godoc.org/github.com/jpillora/go-ogle-analytics
 */
func InitializeAnalytics(appName string, appVersion string) {
	analyticsEnabled = true

	// Locale
	var systemLocale, systemLocaleErr = GetSystemLocale()
	if systemLocaleErr != nil {
		// set to default value if discovery failed
		systemLocale = "XX_XX"
	}

	// Initialize Client
	segmentClient, segmentClientErr = segment.NewWithConfig("6B5KsCRcmam7tIFqpVqEStod6QAo4Ttp", segment.Config{
		BatchSize: 1,
		DefaultContext: &segment.Context{
			App: segment.AppInfo{
				Name:    appName,
				Version: appVersion,
			},
			Locale: systemLocale,
		},
	})

	// In case of error disable analytics to keep the tool working
	if segmentClientErr != nil {
		sentry.HandleError(segmentClientErr)
		analyticsEnabled = false
	}

	// Unique User Id
	deviceid, deviceidErr := machineid.ProtectedID(appName)
	if deviceidErr != nil {
		sentry.HandleError(deviceidErr)
	} else {
		uniqueId = deviceid
	}

	// Platform
	var platform = "Desktop"
	if util.IsCIEnvironment() {
		platform = "CI"
	}

	// Segment Info
	segmentClient.Enqueue(segment.Identify{
		UserId: uniqueId,
		Traits: segment.NewTraits().
			SetName(uniqueId).
			Set("hostname", GetHostname()).
			Set("os", runtime.GOOS).
			Set("cpu_cores", strconv.Itoa(runtime.NumCPU())).
			Set("platform", platform).
			Set("locale", systemLocale),
	})
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

/**
 * CleanUp makes sure, that all events have been trasmittet prior to ending the process.
 */
func CleanUp() {
	if analyticsEnabled == true && segmentClient != nil {
		segmentClient.Close()
	}
}
