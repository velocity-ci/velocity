#!/bin/sh -e


statik -src=configs/sql/migrations -dest pkg/grpc/architect/db -p migrations
