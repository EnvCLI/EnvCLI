package sentry

import (
	raven "github.com/getsentry/raven-go"
)

// Run docker instance
func InitializeSentryIO(appVersion string) {
	// Initialize SentryIO
	raven.SetDSN("https://1721bccab8fa4ddbbeb62923fea0d12f:580013da54814c71ab4dcdf6b61edd9f@sentry.io/1407071")
	raven.SetDefaultLoggerName("log")
	raven.SetEnvironment("production")
	raven.SetRelease(appVersion)
	raven.SetIncludePaths([]string{"/github.com/EnvCLI/EnvCLI"})
}

func HandleError(err error) {
	raven.CaptureErrorAndWait(err, nil)
}
