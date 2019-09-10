package container_runtime

import (
	"testing"
)

func TestSanitizeCommandPowerShell(t *testing.T) {
	command := sanitizeCommand("powershell", "command string")
	if command != "powershell command string" {
		t.Errorf("Powershell parsing error")
	}
}

func TestSanitizeCommandSh(t *testing.T) {
	command := sanitizeCommand("sh", "command string")
	if command != "\"/usr/bin/env\" \"sh\" \"-c\" \"command string\"" {
		t.Errorf("Command Sh error %s", command)
	}
}

func TestSanitizeCommandBash(t *testing.T) {
	command := sanitizeCommand("bash", "command string")
	if command != "\"/usr/bin/env\" \"bash\" \"-l\" \"-c\" \"command string\"" {
		t.Errorf("Command bash error %s", command)
	}
}

func TestSystemExec(t *testing.T) {
	err := systemExec("ls")
	if err != nil {
		t.Errorf("Error %s", err)
	}
}
