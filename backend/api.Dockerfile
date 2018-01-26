FROM golang:1.9
ENV CGO_ENABLED=1
RUN go get -u github.com/golang/dep/cmd/dep
RUN go get -u github.com/mattn/go-sqlite3 && \ 
    go install github.com/mattn/go-sqlite3
WORKDIR /go/src/github.com/velocity-ci/velocity/backend
COPY . .
RUN dep ensure -v
RUN GOOS=linux go build -a -installsuffix cgo -o dist/velocity_api ./api

FROM alpine

RUN apk --no-cache --update add ca-certificates sqlite-libs

ENV JWT_SECRET changeme
ENV PORT 80

COPY --from=0 /go/src/github.com/velocity-ci/velocity/backend/dist/velocity_api /bin/velocity_api
ENTRYPOINT ["/bin/velocity_api"]
