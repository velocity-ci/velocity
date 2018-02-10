#!/bin/sh -e

GOOS=linux go build -a -installsuffix cgo -o dist/vci-architect ./cmd/vci-architect