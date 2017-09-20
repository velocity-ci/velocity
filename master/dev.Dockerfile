FROM golang

ENV JWT_SECRET changeme
ENV PORT 80

RUN curl https://glide.sh/get | sh

RUN go get github.com/pilu/fresh
RUN go get github.com/golang/lint/golint

COPY docker/setup_ssh.sh /bin/setup_ssh.sh
COPY docker/start_dev.sh /bin/start_dev.sh
RUN chmod +x /bin/*.sh
