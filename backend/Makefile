GIT_VERSION = $(shell git describe --always)
AWS_DEFAULT_REGION ?= eu-west-1

dep-install:
	@docker run --rm -it \
	--volume ${CURDIR}:/go/src/github.com/velocity/velocity-ci/backend \
	--workdir /go/src/github.com/velocity/velocity-ci/backend \
	golang:1.9 \
	scripts/install-deps.sh

build-architect:
	docker build -t civelocity/architect:${GIT_VERSION} -f build/docker/architect.Dockerfile .
	docker tag civelocity/architect:${GIT_VERSION} civelocity/architect:latest

build-builder:
	docker build -t civelocity/architect:${GIT_VERSION} -f build/docker/builder.Dockerfile .
	docker tag civelocity/builder:${GIT_VERSION} civelocity/builder:latest

build-cli:
	@docker run --rm -it \
	--volume ${CURDIR}:/go/src/github.com/velocity/velocity-ci/backend \
	--workdir /go/src/github.com/velocity/velocity-ci/backend \
	golang:1.9 \
	scripts/build-cli.sh

build-all: build-architect build-builder build-cli

test:
	docker-compose -f docker-compose.test.yml down
	docker-compose -f docker-compose.test.yml up --abort-on-container-exit

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
