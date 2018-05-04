#!/bin/sh -e

go get -u github.com/mitchellh/gox

gox -output="dist/{{.Dir}}_{{.OS}}_{{.Arch}}" \
    -osarch="darwin/386 darwin/amd64 linux/386 linux/amd64 linux/arm" \
    ./cmd/vcli
