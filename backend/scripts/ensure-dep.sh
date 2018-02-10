#!/bin/sh -e

# Golang dep should hopefully be merged in to Go by 1.10 (https://github.com/golang/dep/wiki/Roadmap)
if ! [ -x "$(command -v dep)" ]; then
    echo "dep is not installed. Installing..."
    dep_version=0.4.1
    curl -fsSL -o /usr/local/bin/dep \
    https://github.com/golang/dep/releases/download/v${dep_version}/dep-linux-amd64 && \
    chmod +x /usr/local/bin/dep
fi
echo "dep is installed."
