+++
date = "2012-08-15T22:32:09+01:00"
title = "Docker Compose"
+++

This guide will show you how to get quickly up and running with Docker Compose locally.

{{< highlight yaml >}}

version: '3'
  services:

    architect:
      image: civelocity/architect:latest
      environment:
        ADMIN_PASSWORD: velocity_local1234
        JWT_SECRET: jwt_local1234
        BUILDER_SECRET: builder_secret1234
      ports:
      - "80:80"
      volumes:
      - "./architect_data:/opt/velocityci"

    builder:
      image: civelocity/builder:latest
      environment:
        BUILDER_SECRET: builder_secret1234
        ARCHITECT_ADDRESS: http://architect
      volumes:
       - "/opt/velocityci:/opt/velocityci"
       - "/var/run/docker.sock:/var/run/docker.sock"

    web:
      image: civelocity/web:latest
      environment:
        ARCHITECT_ENDPOINT="http://localhost/v1"
      ports:
      - "4200:80"

{{< / highlight >}}
