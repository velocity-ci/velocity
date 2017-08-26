GIT_VERSION = $(shell git describe --always)

build_dev:
	docker build -f dev.Dockerfile -t velocity-dev .

build: build_dev

	docker run --rm \
	--volume ${CURDIR}:/app \
	--workdir /app/velocity \
	--env TERM=xterm \
	--env MIX_ENV=prod \
	velocity-dev \
	/bin/sh -c "mix deps.get && mix deps.compile && mix release --env=prod --verbose"

	docker build -f app.Dockerfile -t vjftw/velocity:master-${GIT_VERSION} -t vjftw/velocity:master .