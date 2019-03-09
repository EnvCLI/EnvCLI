[![Build Status](https://travis-ci.org/PhilippHeuer/EnvCLI.svg?branch=master)](https://travis-ci.org/PhilippHeuer/EnvCLI)
[![Go Report Card](https://goreportcard.com/badge/philippheuer/envcli)](http://goreportcard.com/report/philippheuer/envcli)
[![Maintenance](https://img.shields.io/maintenance/yes/2018.svg)]()
[![GitHub contributors](https://img.shields.io/github/contributors/PhilippHeuer/envcli.svg)]()
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/PhilippHeuer/envcli/blob/master/LICENSE.md)
[![Version](https://img.shields.io/github/tag/philippheuer/envcli.svg)]()

*EnvCLI* is a simple wrapper that allows you to run commands within *ethereal docker containers*. You can configure commands to run in docker images within the envcli configuration file.
It currently supports the following providers: [Docker for Windows](https://docs.docker.com/docker-for-windows/install/), [Docker on Linux](https://docs.docker.com/engine/installation/) and [Docker Toolbox](https://docs.docker.com/toolbox/overview/).

[![Discord Server](https://discordapp.com/api/guilds/488047700764393505/embed.png?style=banner2)](https://discord.gg/9m8tWns)

**What merits does this have?**

- Reproducible builds (every developer and ci always use the specified version of Node, Go, ...)
- Official Docker Image for CI
- You don't have to build docker images with node & angular cli or golang & dep - just use the official containers directly
- Quick on-boarding (just install Docker and EnvCLI and you can start coding without setting up any other dependencies or spending time on configurations)
- Enforce identical development environments (every developer has the same version of the compilers/gradle/...)
- Never install dependencies manually or deal with leftovers of old versions (containers are ethereal)
- Tools (ex. Ruby -> Changelog generator) can be defined in the `.envcli.yml` without installing Ruby or a specific version which might break other tools
- Supports a system-scoped configuration to define tools you need like for example htop, grep, ct (coreos config transpiler), overcommit or other tools
- Suppports the creation of aliases, just use `npm install x` like you usually do and you won't even notice your commands are executed within docker containers by envcli

---

. **[Overview](#overview)** . **[Installation](#installation)** . **[Documentation](#documentation)** . **[Credits](#credits)** .

---

## Overview

To use *EnvCLI* you have to install docker and envcli. (See **[Installation](#installation)**)

After that you can create the `.envcli.yml` configuration file for your project.

Example (A single image can provide multiple commands):
```
commands:
- name: npm
  description: Node.js is a JavaScript-based platform for server-side and networking applications.
  provides:
  - npm
  - yarn
  image: docker.io/node:10-alpine
```

When you run `envcli run npm init` *EnvCLI* will take the executed command and match it to the [Docker](https://www.docker.com/) Image `node:10-alpine` based on the provided commands.

You can also use `envcli install-aliases --scope project` to install the project defined aliases and use `npm init` directly - envcli will create a script in your path that will redirect your command to envcli and cause it to be executed within a container.

#### What does EnvCLI do?

This project only provides the configuration file and the easy *envcli* commmand, therefore making it easier to use [Docker](https://www.docker.com/) when development your project. You can do the same without *EnvCLI*.

**Plain Docker**:
```
docker run --rm -it --workdir /go/src/project/ --volume "C:\SourceCodes\golang\envcli:/
go/src/project" golang:latest /usr/bin/env sh -c "go build -o envcli src/*"
```

**With EnvCLI**:
```
envcli run go build -o envcli src/*
```

## Installation

#### **Docker for Windows**

1. Install Docker for Windows from https://docs.docker.com/docker-for-windows/install/
2. Install [EnvCLI](https://dl.bintray.com/envcli/golang/envcli/v0.4.0/EnvCLI-amd64.msi)

#### **Docker for Linux**

1. Install the default Docker version from your favorite package manager.
2. Install [EnvCLI]

*32bit*
```
$ curl -L -o /usr/local/bin/envcli https://dl.bintray.com/envcli/golang/envcli/v0.4.0/envcli_linux_386
$ chmod +x /usr/local/bin/envcli
```

*64bit*
```
$ curl -L -o /usr/local/bin/envcli https://dl.bintray.com/envcli/golang/envcli/v0.4.0/envcli_linux_amd64
$ chmod +x /usr/local/bin/envcli
```

#### **Docker Toolbox (Legacy)**

Install [EnvCLI](https://dl.bintray.com/envcli/golang/envcli/v0.4.0/EnvCLI-amd64.msi)

Now you have to configure a docker-machine for envcli: `docker-machine create envcli`

After that you have to share the drive containing your projects with virtualbox and docker:

 1. Stop the envcli machine `docker-machine stop envcli`
 2. Open VirtualBox
 3. Rightclick -> Settings on the envcli virtual Machine
 4. Select Shared Folders and add a machine folder with the Path
 5. Share the drive which contains your projects (In this example C -> Folder_Path: `C:\`, Folder_Name: `C_DRIVE`) and select the options `Auto-mount` and `Permanent`
 6. Start the envcli machine `docker-machine start envcli`

## Roadmap

- [Features](https://github.com/PhilippHeuer/EnvCLI/labels/feature)

## Documentation

- [Documentation](https://envcli.readthedocs.io/en/latest/)
- [Examples](https://envcli.readthedocs.io/en/latest/examples/)
- [Changelog](https://envcli.readthedocs.io/en/latest/changelog/overview/)

## Credits

- [Bintray - Software Distribution](https://bintray.com)
- [Advanced Installer](https://www.advancedinstaller.com/) - Free License to build the Setup
- [Jinzhu / YML Configuration File](https://github.com/jinzhu/configor)
- [Sirupsen / Logging](https://github.com/sirupsen/logrus)
- [Urfave / CLI](https://github.com/urfave/cli)
- [mattn / go-isatty](https://github.com/mattn/go-isatty)
- [Mattn / Support of colors in Windows CLI](https://github.com/mattn/go-colorable)
- [Inconshreveable / Go Update](https://github.com/inconshreveable/go-update)
- [Blang / Semver](https://github.com/blang/semver)
