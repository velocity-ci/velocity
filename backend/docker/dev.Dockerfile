FROM golang:1.9

ENV JWT_SECRET changeme
ENV PORT 80

ENV DEP_VERSION 0.4.1

RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v${DEP_VERSION}/dep-linux-amd64 && chmod +x /usr/local/bin/dep
RUN go get github.com/smartystreets/goconvey
