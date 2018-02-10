FROM golang:1.9
ENV CGO_ENABLED=0
ENV DEP_VERSION 0.4.1

RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v${DEP_VERSION}/dep-linux-amd64 && chmod +x /usr/local/bin/dep

WORKDIR /go/src/github.com/velocity-ci/velocity/backend
COPY . .
RUN dep ensure -v
RUN GOOS=linux go build -a -installsuffix cgo -o dist/vci-architect ./cmd/vci-architect

FROM alpine

RUN apk --no-cache --update add ca-certificates

ENV JWT_SECRET changeme
ENV PORT 80

COPY --from=0 /go/src/github.com/velocity-ci/velocity/backend/dist/vci-architect /bin/vci-architect
ENTRYPOINT ["/bin/vci-architect"]