GIT_VERSION = $(shell git describe --always)
AWS_DEFAULT_REGION ?= eu-west-1
GOPATH ?= ${CURDIR}/.gopath

.PHONY: test install

install:
	@docker run --rm -it \
	--volume ${CURDIR}:/app \
	--volume ${GOPATH}/pkg/mod/cache:/go/pkg/mod/cache \
	--workdir /app \
	--env GO111MODULE=on \
	golang:1.12 \
	scripts/install-deps.sh

test:
	@docker run --rm -it \
	--volume ${CURDIR}:/app \
	--volume ${GOPATH}/pkg/mod/cache:/go/pkg/mod/cache \
	--workdir /app \
	--env GO111MODULE=on \
	golang:1.12 \
	scripts/test-unit.sh

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
