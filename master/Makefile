GIT_VERSION = $(shell git describe --always)

update:
	docker run --rm \
	--volume ${CURDIR}:/go/src/github.com/VJftw/velocity/master/ \
	--workdir /go/src/github.com/VJftw/velocity/master/velocity \
 	--env CGO_ENABLED=0 \
 	--env GOOS=linux \
	golang \
	/bin/sh -c "curl https://glide.sh/get | sh && glide update"

install:
	docker run --rm \
	--volume ${CURDIR}:/go/src/github.com/VJftw/velocity/master/ \
	--workdir /go/src/github.com/VJftw/velocity/master/velocity \
 	--env CGO_ENABLED=0 \
 	--env GOOS=linux \
	golang \
	/bin/sh -c "curl https://glide.sh/get | sh && glide install"

build: install
	docker run --rm \
	--volume ${CURDIR}:/go/src/github.com/VJftw/velocity/master/ \
	--workdir /go/src/github.com/VJftw/velocity/master/velocity \
 	--env CGO_ENABLED=0 \
 	--env GOOS=linux \
	golang \
	/bin/sh -c "go build -a -installsuffix cgo -o dist/velocity"

	docker build -t vjftw/velocity:master-${GIT_VERSION} .
	docker tag vjftw/velocity:master-${GIT_VERSION} vjftw/velocity:master-latest

build-dev:
	docker build -t vjftw/velocity:master-dev -f dev.Dockerfile .

test: build-dev install
	$(eval NOVENDOR=$(shell docker run --rm \
	--volume ${CURDIR}:/go/src/github.com/VJftw/velocity/master/ \
	--workdir /go/src/github.com/VJftw/velocity/master/velocity \
	vjftw/velocity:master-dev \
	glide novendor))

	docker run --rm \
	--volume ${CURDIR}:/go/src/github.com/VJftw/velocity/master/ \
	--workdir /go/src/github.com/VJftw/velocity/master/velocity \
	vjftw/velocity:master-dev \
	go test -v -cover ${NOVENDOR}

	docker run --rm \
    	--volume ${CURDIR}:/go/src/github.com/VJftw/velocity/master/ \
    	--workdir /go/src/github.com/VJftw/velocity/master/velocity \
    	vjftw/velocity:master-dev \
    	golint -set_exit_status ${NOVENDOR}