package container_runtime

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestSetName(t *testing.T) {
	container := Container{}
	container.SetName("testCase")

	containerCmd := container.GetRunCommand()
	if strings.Contains(containerCmd, "--name \"testCase\"") == false {
		t.Errorf("--name not set correctly")
	}
}

func TestSetEntrypoint(t *testing.T) {
	container := Container{}
	container.SetEntrypoint("/bin/test")

	containerCmd := container.GetRunCommand()
	if strings.Contains(containerCmd, "--entrypoint \"/bin/test\"") == false {
		t.Errorf("--entrypoint not set correctly")
	}
}

func TestSetEnvironment(t *testing.T) {
	container := Container{}
	container.AddEnvironmentVariable("DEBUG", "true")

	containerCmd := container.GetRunCommand()
	if strings.Contains(containerCmd, fmt.Sprintf("-e %s=%s", "DEBUG", strconv.Quote("true"))) == false {
		t.Errorf("-e not set correctly")
	}
}

func TestPublishPort(t *testing.T) {
	container := Container{}
	container.AddContainerPort(ContainerPort{Source: 80, Target: 80})

	containerCmd := container.GetRunCommand()
	if strings.Contains(containerCmd, fmt.Sprintf("-p %d:%d", 80, 80)) == false {
		t.Errorf("-p not set correctly")
	}
}

func TestSetWorkingDirectory(t *testing.T) {
	container := Container{}
	container.SetWorkingDirectory("/home/app")

	containerCmd := container.GetRunCommand()
	if strings.Contains(containerCmd, fmt.Sprintf("--workdir %s", strconv.Quote("/home/app"))) == false {
		t.Errorf("--workdir not set correctly")
	}
}

func TestAddVolume(t *testing.T) {
	container := Container{}
	container.AddVolume(ContainerMount{MountType: "directory", Source: "/root", Target: "/root"})

	containerCmd := container.GetRunCommand()
	if strings.Contains(containerCmd, "-v \"/root:/root\"") == false {
		t.Errorf("-v volume not set correctly")
	}
}

func TestSetImage(t *testing.T) {
	container := Container{}
	container.SetImage("alpine:latest")

	containerCmd := container.GetRunCommand()
	if strings.Contains(containerCmd, "alpine:latest") == false {
		t.Errorf("image not set correctly")
	}
}

func TestSetCommand(t *testing.T) {
	container := Container{}
	container.SetCommandShell("sh")
	container.SetCommand("printenv")

	containerCmd := container.GetRunCommand()
	if strings.HasSuffix(containerCmd, "/usr/bin/env sh -c \"printenv\"") == false {
		t.Errorf("command not set correctly")
	}
}
