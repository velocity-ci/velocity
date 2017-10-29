#!/bin/bash

# Download script to ~/.bin
mkdir -p ~/.bin
curl -L https://raw.githubusercontent.com/velocity-ci/velocity/master/cli/velocity.sh -o ~/.bin/velocity
chmod +x ~/.bin/velocity

# Update path
echo $PATH | grep ~/.bin >/dev/null || (PATH=$PATH:~/.bin && echo "" && echo "Please add ~/.bin to your PATH to make this permanent" && echo "export PATH=\$PATH:~/.bin" && echo "")

# Download latest cli
docker pull civelocity/cli:latest
