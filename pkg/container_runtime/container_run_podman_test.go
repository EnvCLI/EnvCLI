package container_runtime

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
)

func TestPodmanSetParamsInteractive(t *testing.T) {
	container := Container{}
	_ = os.Unsetenv("CI")

	containerCmd := container.GetRunCommand("podman")
	assert.Contains(t, containerCmd, "-ti", "params should include -ti")
}

func TestPodmanSetParamsCI(t *testing.T) {
	container := Container{}
	_ = os.Setenv("CI", "true")

	containerCmd := container.GetRunCommand("podman")
	assert.NotContains(t, containerCmd, "-ti", "params should not include -ti")
}

func TestPodmanSetName(t *testing.T) {
	container := Container{}
	container.SetName("testCase")

	containerCmd := container.GetRunCommand("podman")
	assert.Contains(t, containerCmd, "--name \"testCase\"", "name not set correctly")
}

func TestPodmanSetEntrypoint(t *testing.T) {
	container := Container{}
	container.SetEntrypoint("/bin/test")

	containerCmd := container.GetRunCommand("podman")
	assert.Contains(t, containerCmd, "--entrypoint \"/bin/test\"", "entrypoint not set correctly")
}

func TestPodmanSetEnvironment(t *testing.T) {
	container := Container{}
	container.AddEnvironmentVariable("DEBUG", "true")

	containerCmd := container.GetRunCommand("podman")
	assert.Contains(t, containerCmd, fmt.Sprintf("-e %s=%s", "DEBUG", strconv.Quote("true")), "env not set correctly")
}

func TestPodmanPublishPort(t *testing.T) {
	container := Container{}
	container.AddContainerPort(ContainerPort{Source: 80, Target: 80})

	containerCmd := container.GetRunCommand("podman")
	assert.Contains(t, containerCmd, fmt.Sprintf("-p %d:%d", 80, 80), "publish port not set correctly")
}

func TestPodmanSetWorkingDirectory(t *testing.T) {
	container := Container{}
	container.SetWorkingDirectory("/home/app")

	containerCmd := container.GetRunCommand("podman")
	assert.Contains(t, containerCmd, fmt.Sprintf("--workdir %s", strconv.Quote("/home/app")), "workdir not set correctly")
}

func TestPodmanAddVolume(t *testing.T) {
	container := Container{}
	container.AddVolume(ContainerMount{MountType: "directory", Source: "/root", Target: "/root"})

	containerCmd := container.GetRunCommand("podman")
	assert.Contains(t, containerCmd, "-v \"/root:/root\"", "mount not set correctly")
}

func TestPodmanSetUserArgs(t *testing.T) {
	container := Container{}
	container.SetUserArgs("--link hello:world")

	containerCmd := container.GetRunCommand("podman")
	assert.Contains(t, containerCmd, "--link hello:world", "user args nto set correctly")
}

func TestPodmanSetImage(t *testing.T) {
	container := Container{}
	container.SetImage("alpine:latest")

	containerCmd := container.GetRunCommand("podman")
	assert.Contains(t, containerCmd, "alpine:latest", "image not set correctly")
}

func TestPodmanSetCommand(t *testing.T) {
	container := Container{}
	container.SetCommandShell("sh")
	container.SetCommand("printenv")

	containerCmd := container.GetRunCommand("podman")
	assert.Contains(t, containerCmd, "\"/usr/bin/env\" \"sh\" \"-c\" \"printenv\"", "container command invalid")
}
