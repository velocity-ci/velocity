#!/bin/sh -e

if ! [ -x "$(command -v godog)" ]; then
    echo "godog is not installed. Installing..."
    go get github.com/DATA-DOG/godog/cmd/godog
fi
echo "godog is installed."

cd test
godog
