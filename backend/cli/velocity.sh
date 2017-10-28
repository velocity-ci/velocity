#!/bin/sh

docker pull civelocity/cli:latest

docker run --rm \
--volume $(pwd):/app \
--volume /var/run/docker.sock:/var/run/docker.sock \
--workdir /app \
--env SIB_CWD=$(pwd) \
-it \
civelocity/cli:latest $1 $2
