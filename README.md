# EnvCLI

**Work in Progress**

With EnvCLI you only have to install Docker on your Development Machine and define your environment on a per-project basis in the `.envcli.yml` configuration.
All commands will be executed inside of *containers*, so there is no need to install any local dependencies (go, node, ...) or deal with conflicting versions said dependencies.

## Why should i use this?

 - Use project-specific versions of your dependencies (go, node, ...) to build or run your project
 - Enforce identical development environments for all developers
 - Never install dependencies manually or deal with leftovers (containers are used and discarded
 - Package tools (ex. Ruby -> Changelog generator) can be defined in the `.envcli.yml` without installing Ruby or a specific version which might break other tools
 - Supports inputs from the dependencies as in `npm init` and other commands

## How to use it?

#### Example: Build a Go-Project
```
$ cd /myproject
$ envcli run go build *.go
INFO[0000] Redirecting command to Docker Container [golang:latest].
Unable to find image 'golang:latest' locally
latest: Pulling from library/golang
723254a2c089: Pull complete
abe15a44e12f: Pull complete
409a28e3cc3d: Pull complete
503166935590: Pull complete
abe52c89597f: Pull complete
ce145c5cf4da: Pull complete
96e333289084: Pull complete
39cd5f38ffb8: Pull complete
Digest: sha256:9ccc9da90832f1d48ea379be18700a92e8274efdfb0891d3385f314fc6574976
Status: Downloaded newer image for golang:latest
```
This also means that you can have your go projects wherever you want, since your project will be mounted within the `go path` by default.

## Installation

#### **Docker for Windows**

1. Install Docker for Windows from https://docs.docker.com/docker-for-windows/install/
2. Install [EnvCLI](https://bin.equinox.io/c/ezRXCVys3aV/envcli-stable-windows-amd64.msi)

#### **Docker for Linux**

1. Install the default Docker version from your favorite package manager.
2. Install [EnvCLI](https://dl.equinox.io/philippheuer/envcli/stable)

#### **Docker Toolbox (Legacy)**

The first step is to create a new docker-machine which will be used by envcli: `docker-machine create envcli`

After that you have to share the drive containing your projects with virtualbox and docker:
 1. Stop the envcli machine `docker-machine stop envcli`
 2. Open VirtualBox
 3. Rightclick -> Settings on the envcli virtual Machine
 4. Select Shared Folders and add a machine folder with the Path
 5. Share the drive which contains your projects (In this example C -> Folder_Path: `C:\`, Folder_Name: `C_DRIVE`) and select the options `Auto-mount` and `Permanent`
 6. Start the envcli machine `docker-machine start envcli`

## Credits

- YML Configuration File [github.com/jinzhu/configor]
- Logging [github.com/sirupsen/logrus]
- CLI [github.com/urfave/cli]
- Support of colors in Windows CLI [github.com/mattn/go-colorable]
