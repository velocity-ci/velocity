FROM golang:1.9

ENV JWT_SECRET changeme
ENV PORT 80

RUN curl https://glide.sh/get | sh

RUN go get github.com/pilu/fresh
RUN go get github.com/golang/lint/golint
RUN go get github.com/smartystreets/goconvey
