FROM golang:1.9
ENV CGO_ENABLED=0
ENV DEP_VERSION 0.4.1

RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v${DEP_VERSION}/dep-linux-amd64 && chmod +x /usr/local/bin/dep

WORKDIR /go/src/github.com/velocity-ci/velocity/backend
COPY . .
RUN dep ensure -v
RUN GOOS=linux go build -a -installsuffix cgo -o dist/vci-builder ./cmd/vci-builder

FROM alpine

RUN apk --no-cache --update add ca-certificates

ENV BUILDER_SECRET changeme
ENV ARCHITECT_ADDRESS changeme

COPY --from=0 /go/src/github.com/velocity-ci/velocity/backend/dist/vci-builder /bin/vci-builder
ENTRYPOINT ["/bin/vci-builder"]
