
FROM golang:1.9
ENV CGO_ENABLED=1
ENV DEP_VERSION 0.4.1

RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v${DEP_VERSION}/dep-linux-amd64 && chmod +x /usr/local/bin/dep
RUN go get -u github.com/mattn/go-sqlite3 && \ 
    go install github.com/mattn/go-sqlite3
WORKDIR /go/src/github.com/velocity-ci/velocity/backend
COPY . .
RUN dep ensure -v
RUN GOOS=linux go build -a -installsuffix cgo -o dist/velocity_architect ./architect

FROM alpine

RUN apk --no-cache --update add ca-certificates sqlite-libs

ENV JWT_SECRET changeme
ENV PORT 80

COPY --from=0 /go/src/github.com/velocity-ci/velocity/backend/dist/velocity_architect /bin/velocity_architect
ENTRYPOINT ["/bin/velocity_architect"]