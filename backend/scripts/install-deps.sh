#!/bin/sh -e

if [ ! -d "vendor" ]; then
    scripts/ensure-dep.sh

    dep ensure -v
else
    echo "skipping as already vendored"
fi
