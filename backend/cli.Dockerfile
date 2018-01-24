FROM golang:1.9
RUN go get -u github.com/golang/dep/cmd/dep
WORKDIR /go/src/github.com/velocity-ci/velocity/backend
COPY . .
RUN dep ensure -v
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dist/velocity_cli ./cli

FROM scratch

COPY --from=0 /go/src/github.com/velocity-ci/velocity/backend/dist/velocity_cli /bin/velocity_cli
ENTRYPOINT ["/bin/velocity_cli"]
