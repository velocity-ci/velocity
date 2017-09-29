FROM alpine

RUN apk --no-cache --update add ca-certificates

ENV JWT_SECRET changeme
ENV PORT 80

ADD velocity/dist/velocity /velocity

CMD ["/velocity"]