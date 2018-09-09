# Specification of the .EnvCLI.yml

## Format

The EnvCLI Configuration follows the [yml specifcation](http://yaml.org/spec/).

## Content

The `.envcli.yml` only contains a array of `images`.

The image array has the following attributes:

| Attribute        | Description                                      | Example              |
| ---------------- |:------------------------------------------------:| --------------------:|
| name             | Name of the image                                | Git                  |
| description      | What is this image about?                        | Git VCS              |
| provides         | List of commands that this image provides        | git                  |
| image            | Container Image with Tag                         | docker.io/alpine:git |
| cache            | Cache files on the host (for package manager)    |                      |
| before_script    | Run the provided script lines before the command |                      |
