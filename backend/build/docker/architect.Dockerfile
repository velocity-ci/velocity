FROM golang:1.9
ENV CGO_ENABLED=0

WORKDIR /go/src/github.com/velocity-ci/velocity/backend
COPY . .
RUN scripts/install-deps.sh
RUN scripts/build-architect.sh

FROM alpine

RUN apk --no-cache --update add ca-certificates

ENV JWT_SECRET changeme
ENV PORT 80

COPY --from=0 /go/src/github.com/velocity-ci/velocity/backend/dist/vci-architect /bin/vci-architect
ENTRYPOINT ["/bin/vci-architect"]