FROM golang:1.9
RUN go get -u github.com/golang/dep/cmd/dep
WORKDIR /go/src/github.com/velocity-ci/velocity/backend
COPY . .
RUN dep ensure -v
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dist/velocity_slave ./slave

FROM alpine

RUN apk --no-cache --update add ca-certificates

ENV MASTER_ADDRESS changeme
ENV SLAVE_SECRET changeme

COPY --from=0 /go/src/github.com/velocity-ci/velocity/backend/dist/velocity_slave /bin/velocity_slave
ENTRYPOINT ["/bin/velocity_slave"]
