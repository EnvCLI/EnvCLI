package analytic

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

/**
 * Get System Locale
 */
func GetSystemLocale() (string, error) {
	// Check the LANG environment variable, common on UNIX.
	envlang, ok := os.LookupEnv("LANG")
	if ok {
		return strings.Split(envlang, ".")[0], nil
	}

	// Exec powershell Get-Culture on Windows.
	cmd := exec.Command("powershell", "Get-Culture | select -exp Name")
	output, err := cmd.Output()
	if err == nil {
		return strings.Trim(string(output), "\r\n"), nil
	}

	return "", fmt.Errorf("cannot determine locale")
}
