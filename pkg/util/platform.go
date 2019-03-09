package util

import "os"

/**
 * Detect CI
 */
func IsCIEnvironment() (val bool) {
	value, found := os.LookupEnv("CI")
	if found && value == "true" {
		return true
	}

	return false
}
