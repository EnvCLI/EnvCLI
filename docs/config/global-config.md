# Global Configuration

It's possible to configure EnvCLI to support global commands you may want to use across projects, but a project can overwrite commands that are defined globally so you don't have to worry about your global configuration messing with the project config.


The global configuration is the same as the `.envcli.yml` within a normal project, you just need to set the path in which EnvCLI should search for your custom configuration with `envcli config set global-configuration-path S:\Configuration\EnvCLI`.

This is a sample file that covers a few different use-cases and shows multiple commands by one image, caching and a few other features.

If you'r looking for the specification of the `.envcli.yml` file take a look at the project config page.

```
commands:
# General
- name: alpine
  description: Alpine Linux is a Linux distribution based on musl and BusyBox, primarily designed for "power users who appreciate security, simplicity and resource efficiency".
  provides:
  - ls
  image: docker.io/alpine:latest
  shell: sh
# Development
# - General
- name: Git
  description: Git VCS
  provides:
  - git
  image: docker.io/alpine:git
# - Web Development
- name: npm
  description: Node.js is a JavaScript-based platform for server-side and networking applications.
  provides:
  - npm
  - yarn
  image: docker.io/node:10-alpine
  cache:
  - name: node-10
    directory: /root/.npm
# Clients
# - Cloud
- name: Google Cloud SDK
  description: Google Cloud SDK bundle with all components and dependencies 
  provides:
  - gcloud
  image: docker.io/google/cloud-sdk:alpine
# Infrastructure
# - CoreOS
- name: CoreOS Configuration Transpiler
  description: Ignition is a new provisioning utility designed specifically for CoreOS Container Linux.
  provides:
  - ct
  image: docker.io/envcli/coreos-ignition-configuration-transpiler:latest
  shell: sh
  before_script:
  - echo Hello World
# - Kubernetes
- name: Helm Client
  description: Helm is a tool for managing Kubernetes charts. Charts are packages of pre-configured Kubernetes resources.
  provides:
  - helm
  image: docker.io/linkyard/docker-helm:2.10.0
  shell: sh
```