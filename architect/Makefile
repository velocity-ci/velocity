GIT_VERSION = $(shell git describe --always)
AWS_DEFAULT_REGION ?= eu-west-1

.PHONY: test install

install:
	@docker run --rm -it \
	--volume ${CURDIR}:/app \
	--workdir /app \
	elixir:latest \
	/bin/sh -c "mix local.hex --force && mix deps.get"

test:
	@docker-compose -f docker-compose.test.ymal down || true
	@docker-compose -f docker-compose.test.yaml up --abort-on-container-exit
