package analytic

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
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
	if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "Get-Culture | select -exp Name")
		output, err := cmd.Output()
		if err == nil {
			return strings.Trim(string(output), "\r\n"), nil
		}
	}

	return "", fmt.Errorf("cannot determine locale")
}

/**
 * GetHostname gets the hostname ...
 */
func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}

	addrs, err := net.LookupIP(hostname)
	if err != nil {
		return hostname
	}

	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			ip, err := ipv4.MarshalText()
			if err != nil {
				return hostname
			}
			hosts, err := net.LookupAddr(string(ip))
			if err != nil || len(hosts) == 0 {
				return hostname
			}
			fqdn := hosts[0]
			return strings.TrimSuffix(fqdn, ".")
		}
	}

	return hostname
}
