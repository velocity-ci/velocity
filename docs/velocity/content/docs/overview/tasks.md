+++
date = "2012-08-15T22:32:09+01:00"
title = "Tasks"
+++

# Overview

Tasks are configured in [YAML](http://yaml.org/) and consist of *Steps*. There are different *Step* types (see below).

The simplest possible Task we can define is:


```
# hello-velocity.yml
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
```
vcli run hello-velocity
```

# Steps

The following *Steps* should suit all needs CI/CD & task running needs.

## Docker Build

## Docker Run

## Docker Compose

## Plugin

## Push

