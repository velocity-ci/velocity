---
sidebarDepth: 3
---

# Overview

## Tasks

Tasks are configured in [YAML](http://yaml.org/) and consist of _Steps_. There are different _Step_ types (see below).

The simplest possible Task we can define is:

```yaml
# ./tasks/hello-velocity.yml
---
description: "Hello Velocity"
name: hello-velocity

steps:
  - type: run
    description: Hello Docker
    image: hello-world:latest
```

This defines a task with the name `hello-velocity` that will pull and run the `hello-world:latest` docker image, which is the equivalent of doing `docker run --rm hello-world:latest`.

You can run this using the Velocity CLI with:

```bash
vcli run hello-velocity
```

### Parameters

You can define parameters for your task with the `parameters` array and use them with `${<parameter name>}` expressions:

#### Basic Parameters

```yaml
# ./tasks/hello-parameters.yml
---
description: "Hello example with parameters"
name: hello-parameters

parameters:
  - name: your_name
  - name: your_secret
    secret: true

steps:
  - type: run
    description: Hello!
    image: busybox:latest
    command: echo "Hello ${your_name}. I know your secret ${your_secret}."
```

The above will require the user to enter 2 parameters, `your_name` and `your_secret`. Note that the `your_secret` parameter has `secret: true` which tells Velocity to censor any output matching this secret.

Note: You'll notice this isn't entirely fool-proof if you, for example, set `your_secret` to "your", the output of the above task will then look something like:

```
Hello Bob. I know *** secret ***.
```

#### Derived Parameters

Derived parameters run a Go binary that can return any arbitrary information to be used as parameters:

```yaml
# ./tasks/publish-cli.yml
---
description: "Publish the Velocity CLI"
name: publish-cli

parameters:
  - use: https://github.com/velocity-ci/parameter.aws-ssm/releases/download/0.1.1/aws-ssm
    arguments:
      name: /velocityci/github-release-token
    exports:
      value: github_release_token
```

The above example shows use of the [Velocity AWS SSM parameter](https://github.com/velocity-ci/parameter.aws-ssm) binary exporting the `value` of `/velocityci/github-release-token` as `github_release_token`. The `github_release_token` is then used in creating a GitHub release for the CLI of Velocity!

### Steps

The following _Steps_ should suit most (if not all) needs for CI/CD & task running needs.

#### Docker Build

#### Docker Run

::: tip
If you need to run multiple commands, pipe, redirect or use any other shell features; wrap them in a shell command e.g.

```
# for sh
/bin/sh -c 'echo "Hello"; echo "World"'
# or for bash (Bourne-Again SHell)
/bin/bash -c 'echo "Hello"; echo "World"'
```

Note: Make sure the shell you use is installed on the container that you're running.
:::

#### Docker Compose

#### Push

#### Plugin

## Plugins

## .velocity.yaml
