#!/bin/sh -e

GOOS=linux go build -a -installsuffix cgo -o dist/vci-builder ./cmd/vci-builder