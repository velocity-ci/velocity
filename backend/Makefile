GIT_VERSION = $(shell git describe --always)

dep-install:
	docker run -it --rm \
	--volume ${CURDIR}:/go/src/github.com/velocity-ci/velocity/backend/ \
	--workdir /go/src/github.com/velocity-ci/velocity/backend \
 	--env CGO_ENABLED=0 \
 	--env GOOS=linux \
	civelocity/dev \
	/bin/sh -c "dep ensure -v && dep prune -v"

dep-update:
	docker run -it --rm \
	--volume ${CURDIR}:/go/src/github.com/velocity-ci/velocity/backend/ \
	--workdir /go/src/github.com/velocity-ci/velocity/backend \
 	--env CGO_ENABLED=0 \
 	--env GOOS=linux \
	civelocity/dev \
	/bin/sh -c "dep ensure -v -update && dep prune -v"

dep-status:
	docker run -it --rm \
	--volume ${CURDIR}:/go/src/github.com/velocity-ci/velocity/backend/ \
	--workdir /go/src/github.com/velocity-ci/velocity/backend \
 	--env CGO_ENABLED=0 \
 	--env GOOS=linux \
	civelocity/dev \
	/bin/sh -c "dep status"

build-api:
	docker run --rm \
	--volume ${CURDIR}:/go/src/github.com/velocity-ci/velocity/backend/ \
	--workdir /go/src/github.com/velocity-ci/velocity/backend \
 	--env CGO_ENABLED=1 \
 	--env GOOS=linux \
	golang \
	/bin/sh -c "go build -a -installsuffix cgo -o dist/velocity_api ./api"

	docker build -t civelocity/api:${GIT_VERSION} -f api.Dockerfile .
	docker tag civelocity/api:${GIT_VERSION} civelocity/api:latest

build-cli:
	docker run --rm \
	--volume ${CURDIR}:/go/src/github.com/velocity-ci/velocity/backend/ \
	--workdir /go/src/github.com/velocity-ci/velocity/backend \
 	--env CGO_ENABLED=0 \
 	--env GOOS=linux \
	golang \
	/bin/sh -c "go build -a -installsuffix cgo -o dist/velocity_cli ./cli"

	docker build -t civelocity/cli:${GIT_VERSION} -f cli.Dockerfile .
	docker tag civelocity/cli:${GIT_VERSION} civelocity/cli:latest

build-slave:
	docker run --rm \
	--volume ${CURDIR}:/go/src/github.com/velocity-ci/velocity/backend/ \
	--workdir /go/src/github.com/velocity-ci/velocity/backend \
 	--env CGO_ENABLED=0 \
 	--env GOOS=linux \
	golang \
	/bin/sh -c "go build -a -installsuffix cgo -o dist/velocity_slave ./slave"

	docker build -t civelocity/slave:${GIT_VERSION} -f slave.Dockerfile .
	docker tag civelocity/slave:${GIT_VERSION} civelocity/slave:latest

build-all: build-api build-cli build-slave

build-dev:
	docker build -t civelocity/dev -f dev.Dockerfile .

test: build-dev
	$(eval NOVENDOR=$(shell docker run --rm \
	--volume ${CURDIR}:/go/src/github.com/velocity-ci/velocity/backend/ \
	--workdir /go/src/github.com/velocity-ci/velocity/backend \
	velocity-ci/velocity:backend-dev \
	glide novendor))

	docker run --rm \
	--volume ${CURDIR}:/go/src/github.com/velocity-ci/velocity/backend/ \
	--workdir /go/src/github.com/velocity-ci/velocity/backend \
	civelocity/dev \
	go test -v -cover ${NOVENDOR}

lint:
	$(eval NOVENDOR=$(shell docker run --rm \
	--volume ${CURDIR}:/go/src/github.com/velocity-ci/velocity/backend/ \
	--workdir /go/src/github.com/velocity-ci/velocity/backend \
	civelocity/dev \
	glide novendor))

	docker run --rm \
	--volume ${CURDIR}:/go/src/github.com/velocity-ci/velocity/backend/ \
	--workdir /go/src/github.com/velocity-ci/velocity/backend \
	civelocity/dev \
	golint -set_exit_status ${NOVENDOR}