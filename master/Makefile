GIT_VERSION = $(shell git describe --always)

update:
	docker run --rm \
	--volume ${CURDIR}:/go/src/github.com/velocity-ci/velocity/master/ \
	--workdir /go/src/github.com/velocity-ci/velocity/master/velocity \
 	--env CGO_ENABLED=0 \
 	--env GOOS=linux \
	golang \
	/bin/sh -c "curl https://glide.sh/get | sh && glide update"

install:
	docker run --rm \
	--volume ${CURDIR}:/go/src/github.com/velocity-ci/velocity/master/ \
	--workdir /go/src/github.com/velocity-ci/velocity/master/velocity \
 	--env CGO_ENABLED=0 \
 	--env GOOS=linux \
	golang \
	/bin/sh -c "curl https://glide.sh/get | sh && glide install --strip-vendor"

build: install
	docker run --rm \
	--volume ${CURDIR}:/go/src/github.com/velocity-ci/velocity/master/ \
	--workdir /go/src/github.com/velocity-ci/velocity/master/velocity \
 	--env CGO_ENABLED=0 \
 	--env GOOS=linux \
	golang \
	/bin/sh -c "go build -a -installsuffix cgo -o dist/velocity"

	docker build -t velocity-ci/velocity:master-${GIT_VERSION} -f app.Dockerfile .
	docker tag velocity-ci/velocity:master-${GIT_VERSION} velocity-ci/velocity:master-latest

build-dev:
	docker build -t velocity-ci/velocity:master-dev -f dev.Dockerfile .

test: build-dev install
	$(eval NOVENDOR=$(shell docker run --rm \
	--volume ${CURDIR}:/go/src/github.com/velocity-ci/velocity/master/ \
	--workdir /go/src/github.com/velocity-ci/velocity/master/velocity \
	velocity-ci/velocity:master-dev \
	glide novendor))

	docker run --rm \
	--volume ${CURDIR}:/go/src/github.com/velocity-ci/velocity/master/ \
	--workdir /go/src/github.com/velocity-ci/velocity/master/velocity \
	velocity-ci/velocity:master-dev \
	go test -v -cover ${NOVENDOR}

	docker run --rm \
    	--volume ${CURDIR}:/go/src/github.com/velocity-ci/velocity/master/ \
    	--workdir /go/src/github.com/velocity-ci/velocity/master/velocity \
    	velocity-ci/velocity:master-dev \
    	golint -set_exit_status ${NOVENDOR}