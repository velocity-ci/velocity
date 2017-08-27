GIT_VERSION = $(shell git describe --always)

install:
	docker run --rm -v ${CURDIR}/velocity:/app -w /app node:8 yarn upgrade

serve:
	docker run --rm -it -v ${CURDIR}/velocity:/app -p 4200:4200 -w /app node:8 /bin/sh -c '/app/node_modules/.bin/ng serve --host=0.0.0.0 --disable-host-check'
