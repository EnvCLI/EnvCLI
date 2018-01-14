[![Build Status](https://travis-ci.org/PhilippHeuer/EnvCLI.svg?branch=master)](https://travis-ci.org/PhilippHeuer/EnvCLI)
[![Go Report Card](https://goreportcard.com/badge/philippheuer/envcli)](http://goreportcard.com/report/philippheuer/envcli)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/PhilippHeuer/envcli/blob/master/LICENSE.md)
[![Version](https://img.shields.io/github/tag/philippheuer/envcli.svg)]()

*EnvCLI* is a simple wrapper that allows you to run commands within *ethereal docker containers*. You can match commands to specific containers within a configuration file.
It currently supports several docker backends [Docker for Windows](https://docs.docker.com/docker-for-windows/install/), [Docker on Linux](https://docs.docker.com/engine/installation/) and [Docker Toolbox](https://docs.docker.com/toolbox/overview/).
Since all commands run in *ethereal docker containers* you will never have to install dependencies (Ruby, ...) or other cli tools ever again.

---

. **[Overview](#overview)** . **[Merits](#merits)** . **[Example](#example)** . **[Installation](#installation)** .
. **[Contributing](#contributing)** . **[Changelog](#changelog)** . **[Credits](#credits)** .

---

## Overview

To use *EnvCLI* you have to install docker and envcli. (See **[Installation](#installation)**)

Now you can create your `.envcli.yml` configuration file for your project.

Example:
```
commands:
- name: npm
  description: Node.js is a JavaScript-based platform for server-side and networking applications.
  provides:
  - npm
  - yarn
  image: node
  tag: 9.3.0-alpine
```

When you run `envcli run npm init` *EnvCLI* will detect the above entry based on the provided commands and match it against the [Docker](https://www.docker.com/) Image `node:9.3.0-alpine`.

#### What does the EnvCLI do?

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

## Merits

 - Use project-specific versions of your dependencies (go, node, ...) to build or run your project
 - Enforce identical development environments for all developers
 - Never install dependencies manually or deal with leftovers (containers are ethereal)
 - Package tools (ex. Ruby -> Changelog generator) can be defined in the `.envcli.yml` without installing Ruby or a specific version which might break other tools

## Example

#### Example: Build a Go-Project (Windows by overwriting GOOS and GOARCH)
```
$ cd /myproject
$ envcli --env GOOS=windows --env GOARCH=amd64 run go build -o cli.exe src/*
INFO[0000] Executing specified command in Docker Container [golang:latest].
```

Please take note that your project directory can be anywhere on your computer, because your project will be mounted into the `go path` within the container by default.

#### Example: Node - Init

```
$ envcli run npm init
INFO[0000] Executing specified command in Docker Container [node:9.3.0-alpine].
This utility will walk you through creating a package.json file.
It only covers the most common items, and tries to guess sensible defaults.

See `npm help json` for definitive documentation on these fields
and exactly what they do.

Use `npm install <pkg>` afterwards to install a package and
save it as a dependency in the package.json file.

Press ^C at any time to quit.
package name: (project) myproject
version: (1.0.0) 0.1.0
description: Example
entry point: (index.js)
test command:
git repository:
keywords:
author:
license: (ISC)
About to write to /project/package.json:

{
  "name": "myproject",
  "version": "0.1.0",
  "description": "Example",
  "main": "index.js",
  "scripts": {
    "test": "echo \"Error: no test specified\" && exit 1"
  },
  "author": "",
  "license": "ISC"
}

Is this ok? (yes) yes
```

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

## Roadmap

*Feel free to respond in the issue of the feature if you want to work on any point*

- Aliases for Windows/Linux to omit the "envcli run" prefix and use `npm`, `go`, ... directly. #1
- Caching of directories on the host. #2
- Configuration of proxy server for the containers. #3

## Contributing

Feel free to put up a pull request to fix a bug or maybe add a feature.

## Changelog

#### Unreleased (14-01-2018)

* add the ability to pass environment variables into the containers. [Philipp Heuer]
* add global configuration for the proxy server. [Philipp Heuer]
* support to run commands within subdirectories of the project. [Philipp Heuer]
* add the ability for a single image to provide multiple commands. [Philipp Heuer]
* option to wrap container commands in a shell. [Philipp Heuer]
* support for docker toolbox. [Philipp Heuer]
* using go-colorable to support colored output on windows. [Philipp Heuer]
* self-update command to provide updates. [Philipp Heuer]
* support for command errors and inputs. [Philipp Heuer]

## Credits

- YML Configuration File [github.com/jinzhu/configor]
- Logging [github.com/sirupsen/logrus]
- CLI [github.com/urfave/cli]
- Support of colors in Windows CLI [github.com/mattn/go-colorable]
