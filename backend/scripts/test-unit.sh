#!/bin/sh -e

scripts/install-deps.sh

go test ./... -cover
