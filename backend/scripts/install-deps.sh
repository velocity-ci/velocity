#!/bin/sh -e

go mod download
go mod vendor
go mod verify
