FROM golang:1.9
ENV CGO_ENABLED=0

WORKDIR /go/src/github.com/velocity-ci/velocity/backend
COPY . .
RUN scripts/install-deps.sh
RUN scripts/build-builder.sh

FROM alpine

RUN apk --no-cache --update add ca-certificates

ENV BUILDER_SECRET changeme
ENV ARCHITECT_ADDRESS changeme

COPY --from=0 /go/src/github.com/velocity-ci/velocity/backend/dist/vci-builder /bin/vci-builder
ENTRYPOINT ["/bin/vci-builder"]
