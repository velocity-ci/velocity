GIT_VERSION = $(shell git describe --always)

install:
	docker run --rm -v ${CURDIR}/velocity:/app -w /app node:8 yarn upgrade

serve:
	docker run --rm -it -v ${CURDIR}/velocity:/app -p 4200:4200 -w /app node:8 /bin/sh -c '/app/node_modules/.bin/ng serve --host=0.0.0.0 --disable-host-check'

install_elm:
	docker run --rm -v ${CURDIR}/velocity-elm:/app -w /app node:8 yarn upgrade
	docker run --rm -v ${CURDIR}/velocity-elm:/app -w /app node:8 node_modules/.bin/elm-package install -y

serve_elm:
	docker run --rm -it -v ${CURDIR}/velocity-elm:/app -p 4201:4201 -w /app node:8 /bin/sh -c 'npm rebuild node-sass; ./node_modules/.bin/webpack-dev-server --host 0.0.0.0 --port 4201 --colors'