# EnvCLI *Work in Progress*

With EnvCLI you only have to install Docker on your Development Machine and define your environment on a per-project basis in the `.envcli.yml` configuration file.
All commands/builds/tests will be executed inside of *containers*, so there is no need to install any local dependencies or deal with conflicting versions of `node`, `go` or other tools.

Features:

 - Allows you to use project-specific versions of your environment to build your project/run/test your application in a container

## Installation

#### Container Provider

**Docker for Windows**
Install Docker for Windows from https://docs.docker.com/docker-for-windows/install/

**Docker for Linux**
Install the default Docker version from your favorite package manager.

#### EnvCLI

**Work in Progress**

## Supported Container-Provider

- Docker on Windows
- Docker on Linux

## GO Dependencies

- YML Configuration File [github.com/jinzhu/configor]
- Logging [github.com/sirupsen/logrus]
- CLI [github.com/urfave/cli]
