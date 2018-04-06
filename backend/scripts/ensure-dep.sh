#!/bin/sh -e

# Golang dep should hopefully be merged in to Go by 1.10 (https://github.com/golang/dep/wiki/Roadmap)
if ! [ -x "$(command -v dep)" ]; then
    echo "dep is not installed. Installing..."
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
fi
echo "dep is installed."
