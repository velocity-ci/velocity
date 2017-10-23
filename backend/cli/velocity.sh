#!/bin/sh

# docker pull veloci/cli:latest

docker run --rm \
--volume $(pwd):/app \
--volume /var/run/docker.sock:/var/run/docker.sock \
--workdir /app \
--env SIB_CWD=$(pwd) \
-it \
veloci/cli:latest $1 $2
