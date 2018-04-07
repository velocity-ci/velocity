GIT_VERSION = $(shell git describe --always)


install:
	git submodule update --init --recursive
	docker run --rm -v ${CURDIR}/velocity:/app -w /app node:8 yarn install
	docker run --rm -v ${CURDIR}/velocity:/app -w /app node:8 node_modules/.bin/elm-package install -y

serve:
	docker run --rm -it -v \
	${CURDIR}/velocity:/app \
	 -p 4200:4200 \
	 -w /app \
	 node:8 /bin/sh \
	 -c './node_modules/.bin/webpack-dev-server --host 0.0.0.0 --history-api-fallback --port 4200 --colors'

build:
	@echo "Building frontend release"

	@docker run --rm \
	--volume ${CURDIR}/velocity:/app \
	--workdir /app \
	node:8 \
	sh -c "NODE_ENV=prod ARCHITECT_ADDRESS=__ARCHITECT_ADDRESS__ node_modules/.bin/webpack"

	@docker build -t civelocity/web:${GIT_VERSION} -f docker/Dockerfile .
	@docker tag civelocity/web:${GIT_VERSION} civelocity/web:latest

serve_dist: build
	docker run --rm -p 4200:80 -e ARCHITECT_ENDPOINT='http://localhost/v1' \
	civelocity/web:latest

publish:
	@docker push civelocity/web:${GIT_VERSION}

test: install
	docker run --rm -it -v ${CURDIR}/velocity:/app -w /app node:8 /bin/sh -c 'npm rebuild elm --silent; ./node_modules/.bin/elm-test --compiler ./node_modules/.bin/elm-make'

test_watch:
	docker run --rm -it -v ${CURDIR}/velocity:/app -w /app node:8 /bin/sh -c 'npm rebuild elm --silent; ./node_modules/.bin/elm-test --watch --compiler ./node_modules/.bin/elm-make'

