FROM alpine

RUN apk --no-cache --update add ca-certificates

# API only
ENV JWT_SECRET changeme
ENV PORT 80

# Worker only
ENV POSTGRES_HOST changeme
ENV POSTGRES_USER changeme
ENV POSTGRES_DBNAME changeme
ENV POSTGRES_PASSWORD changeme
ENV FACEBOOK_APP_ID changeme
ENV FACEBOOK_APP_SECRET changeme

ADD velocity/dist/velocity /velocity

CMD ["/velocity"]