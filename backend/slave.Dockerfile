FROM alpine

RUN apk --no-cache --update add ca-certificates

ENV MASTER_ADDRESS changeme
ENV SLAVE_SECRET changeme

ADD dist/velocity_slave /velocity

CMD ["/velocity"]
