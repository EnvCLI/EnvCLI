package docker

import (
	"testing"
)

func TestIsCIEnvironment(t *testing.T) {
	val := IsCIEnvironment()
	if val != false {
		t.Errorf("Ci is set")
	}
}

func TestIsDockerNative(t *testing.T) {
	result := IsDockerNative()
	if result != true {
		t.Errorf("Docker is not native")
	}
}

func TestSanitizeCommandPowerShell(t *testing.T) {
	command := sanitizeCommand("powershell", "command string")
	if command != "powershell command string" {
		t.Errorf("Powershell parsing error")
	}
}

func TestSanitizeCommandSh(t *testing.T) {
	command := sanitizeCommand("sh", "command string")
	if command != "/usr/bin/env sh -c \"command string\"" {
		t.Errorf("Command Sh error %s", command)
	}
}

func TestSanitizeCommandBash(t *testing.T) {
	command := sanitizeCommand("bash", "command string")
	if command != "/usr/bin/env bash -c \"command string\" -l" {
		t.Errorf("Command bash error %s", command)
	}
}

func TestSystemExec(t *testing.T) {
	err := systemExec("ls")
	if err != nil {
		t.Errorf("Error %s", err)
	}
}

/* Add for windows...*/
