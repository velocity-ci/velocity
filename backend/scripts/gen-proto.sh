#!/bin/sh -e

srcDir="${PWD}/api/proto/v1"
goDir="${PWD}/pkg/velocity/genproto"

protoc -I="${PWD}" --go_out="plugins=grpc:${goDir}" ${srcDir}/*.proto
