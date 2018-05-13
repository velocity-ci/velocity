FROM golang:1-alpine
ENV CGO_ENABLED=1

RUN apk add --no-cache curl git gcc musl-dev

WORKDIR /go/src/github.com/velocity-ci/velocity/backend
COPY . .
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
RUN scripts/install-deps.sh
RUN scripts/build-builder.sh

FROM alpine

RUN apk --no-cache --update add ca-certificates openssh-client git

ENV BUILDER_SECRET changeme
ENV ARCHITECT_ADDRESS changeme

COPY build/docker/builder/docker-entrypoint.sh /docker-entrypoint.sh

COPY --from=0 /go/src/github.com/velocity-ci/velocity/backend/dist/vci-builder /bin/vci-builder
ENTRYPOINT ["/docker-entrypoint.sh"]
