package container_runtime

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
)

func TestDockerSetParamsInteractive(t *testing.T) {
	container := Container{}
	_ = os.Unsetenv("CI")

	containerCmd := container.GetRunCommand("docker")
	assert.Contains(t, containerCmd, "-ti", "params should include -ti")
}

func TestDockerSetParamsCI(t *testing.T) {
	container := Container{}
	_ = os.Setenv("CI", "true")

	containerCmd := container.GetRunCommand("docker")
	assert.NotContains(t, containerCmd, "-ti", "params should not include -ti")
}

func TestDockerSetName(t *testing.T) {
	container := Container{}
	container.SetName("testCase")

	containerCmd := container.GetRunCommand("docker")
	assert.Contains(t, containerCmd, "--name \"testCase\"", "name not set correctly")
}

func TestDockerSetEntrypoint(t *testing.T) {
	container := Container{}
	container.SetEntrypoint("/bin/test")

	containerCmd := container.GetRunCommand("docker")
	assert.Contains(t, containerCmd, "--entrypoint \"/bin/test\"", "entrypoint not set correctly")
}

func TestDockerSetEnvironment(t *testing.T) {
	container := Container{}
	container.AddEnvironmentVariable("DEBUG", "true")

	containerCmd := container.GetRunCommand("docker")
	assert.Contains(t, containerCmd, fmt.Sprintf("-e %s=%s", "DEBUG", strconv.Quote("true")), "env not set correctly")
}

func TestDockerPublishPort(t *testing.T) {
	container := Container{}
	container.AddContainerPort(ContainerPort{Source: 80, Target: 80})

	containerCmd := container.GetRunCommand("docker")
	assert.Contains(t, containerCmd, fmt.Sprintf("-p %d:%d", 80, 80), "publish port not set correctly")
}

func TestDockerSetWorkingDirectory(t *testing.T) {
	container := Container{}
	container.SetWorkingDirectory("/home/app")

	containerCmd := container.GetRunCommand("docker")
	assert.Contains(t, containerCmd, fmt.Sprintf("--workdir %s", strconv.Quote("/home/app")), "workdir not set correctly")
}

func TestDockerAddVolume(t *testing.T) {
	container := Container{}
	container.AddVolume(ContainerMount{MountType: "directory", Source: "/root", Target: "/root"})

	containerCmd := container.GetRunCommand("docker")
	assert.Contains(t, containerCmd, "-v \"/root:/root\"", "mount not set correctly")
}

func TestDockerSetUserArgs(t *testing.T) {
	container := Container{}
	container.SetUserArgs("--link hello:world")

	containerCmd := container.GetRunCommand("docker")
	assert.Contains(t, containerCmd, "--link hello:world", "user args nto set correctly")
}

func TestDockerSetImage(t *testing.T) {
	container := Container{}
	container.SetImage("alpine:latest")

	containerCmd := container.GetRunCommand("docker")
	assert.Contains(t, containerCmd, "alpine:latest", "image not set correctly")
}

func TestDockerSetCommand(t *testing.T) {
	container := Container{}
	container.SetCommandShell("sh")
	container.SetCommand("printenv")

	containerCmd := container.GetRunCommand("docker")
	assert.Contains(t, containerCmd, "\"/usr/bin/env\" \"sh\" \"-c\" \"printenv\"", "container command invalid")
}
