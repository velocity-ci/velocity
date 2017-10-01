GIT_VERSION = $(shell git describe --always)

install:
	docker run --rm -v ${CURDIR}/velocity:/app -w /app node:8 yarn install
	docker run --rm -v ${CURDIR}/velocity:/app -w /app node:8 node_modules/.bin/elm-package install -y

serve:
	docker run --rm -it -v ${CURDIR}/velocity:/app -p 4200:4200 -w /app node:8 /bin/sh -c 'npm rebuild node-sass --silent; ./node_modules/.bin/webpack-dev-server --host 0.0.0.0 --port 4200 --colors'

test:
	docker run --rm -it -v ${CURDIR}/velocity:/app -w /app node:8 /bin/sh -c 'npm rebuild elm --silent; ./node_modules/.bin/elm-test --compiler ./node_modules/.bin/elm-make'


test_watch:
	docker run --rm -it -v ${CURDIR}/velocity:/app -w /app node:8 /bin/sh -c 'npm rebuild elm --silent; ./node_modules/.bin/elm-test --watch --compiler ./node_modules/.bin/elm-make'