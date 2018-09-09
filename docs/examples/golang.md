# Example - Golang

This example shows how to work with go and dep for dependency management.

Please take note that your project directory can be anywhere on your computer, because your project will be mounted into the `go path` within the container by default.

## Configuration File `.envcli.yml`

```
images:
- name: go
  description: Go (golang) is a general purpose, higher-level, imperative programming language.
  provides:
  - go
  image: golang
  tag: latest
  directory: /go/src/project
  shell: sh
- name: godep
  description: dep is a prototype dependency management tool for Go. It requires Go 1.8 or newer to compile.
  provides:
  - dep
  image: philippheuer/docker-go-dep
  tag: latest
  directory: /go/src/project
  shell: sh
```

## Build a Go-Project (for Windows, by overwriting GOOS and GOARCH)

```
$ cd /myproject
$ envcli --env GOOS=windows --env GOARCH=amd64 run go build -o cli.exe src/*
INFO[0000] Executing specified command in Docker Container [golang:latest].
```
