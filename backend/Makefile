GIT_VERSION = $(shell git describe --always)
AWS_DEFAULT_REGION ?= eu-west-1

dep-install:
	docker run -it --rm \
	--volume ${CURDIR}:/go/src/github.com/velocity-ci/velocity/backend/ \
	--workdir /go/src/github.com/velocity-ci/velocity/backend \
 	--env CGO_ENABLED=0 \
 	--env GOOS=linux \
	golang:1.9 \
	/bin/sh -c "go get -u github.com/golang/dep/cmd/dep && dep ensure -v"

dep-update:
	docker run -it --rm \
	--volume ${CURDIR}:/go/src/github.com/velocity-ci/velocity/backend/ \
	--workdir /go/src/github.com/velocity-ci/velocity/backend \
 	--env CGO_ENABLED=0 \
 	--env GOOS=linux \
	golang:1.9 \
	/bin/sh -c "go get -u github.com/golang/dep/cmd/dep && dep ensure -v -update"

dep-status:
	docker run -it --rm \
	--volume ${CURDIR}:/go/src/github.com/velocity-ci/velocity/backend/ \
	--workdir /go/src/github.com/velocity-ci/velocity/backend \
 	--env CGO_ENABLED=0 \
 	--env GOOS=linux \
	golang:1.9 \
	/bin/sh -c "go get -u github.com/golang/dep/cmd/dep && dep status"

build-api:
	docker run --rm \
	--volume ${CURDIR}:/go/src/github.com/velocity-ci/velocity/backend/ \
	--workdir /go/src/github.com/velocity-ci/velocity/backend \
 	--env CGO_ENABLED=1 \
 	--env GOOS=linux \
	golang \
	/bin/sh -c "go build -a -installsuffix cgo -o dist/velocity_api ./api"

build-api-image:
	docker build -t civelocity/api:${GIT_VERSION} -f api.Dockerfile .
	docker tag civelocity/api:${GIT_VERSION} civelocity/api:latest

build-cli:
	docker run --rm \
	--volume ${CURDIR}:/go/src/github.com/velocity-ci/velocity/backend/ \
	--workdir /go/src/github.com/velocity-ci/velocity/backend \
 	--env CGO_ENABLED=0 \
 	--env GOOS=linux \
	golang \
	/bin/sh -c 'go get github.com/mitchellh/gox && gox -output="dist/{{.Dir}}_{{.OS}}_{{.Arch}}" -osarch="darwin/386 darwin/amd64 linux/386 linux/amd64 linux/arm" ./cli'

build-cli-image:
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

build-slave-image:
	docker build -t civelocity/slave:${GIT_VERSION} -f slave.Dockerfile .
	docker tag civelocity/slave:${GIT_VERSION} civelocity/slave:latest

build-all: build-api build-cli build-slave

test: download-test-repo
	docker-compose -f docker-compose.test.yml down
	docker-compose -f docker-compose.test.yml up --abort-on-container-exit

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

download-test-repo:
	curl -L https://s3-eu-west-1.amazonaws.com/velocityci.data/test/velocity-gogs-test-data.tar.gz > velocity-gogs-test-data.tar.gz
	tar -xvzf velocity-gogs-test-data.tar.gz

update-test-repo:
	sudo tar -cvzf velocity-gogs-test-data.tar.gz gogs-data
	docker run --rm \
	--volume ${CURDIR}:/app \
	--workdir /app \
	--env AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} \
	--env AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
	--env AWS_DEFAULT_REGION=${AWS_DEFAULT_REGION} \
	anigeo/awscli:latest \
	s3 cp velocity-gogs-test-data.tar.gz s3://velocityci.data/test/velocity-gogs-test-data.tar.gz --acl public-read --cache-control max-age=120
	sudo rm velocity-gogs-test-data.tar.gz
