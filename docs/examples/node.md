# Example - Node

## Configuration File

```
images:
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
```

## Execution

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
