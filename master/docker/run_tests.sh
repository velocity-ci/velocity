#!/bin/sh
set +x
set -e

mix deps.get

/bin/wait-for-it.sh -t 120 postgres:5432 

mix test --color --trace

mix coveralls.html --color

mix credo --strict

mix dogma
