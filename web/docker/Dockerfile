FROM nginx:alpine

ADD https://github.com/just-containers/s6-overlay/releases/download/v1.21.4.0/s6-overlay-amd64.tar.gz /tmp/
RUN tar xzf /tmp/s6-overlay-amd64.tar.gz -C /
ENTRYPOINT ["/init"]

COPY docker/nginx.sh /etc/services.d/nginx/run
RUN chmod 554 /etc/services.d/nginx/run

COPY docker/init.sh /etc/cont-init.d/set_architect
RUN chmod 554 /etc/cont-init.d/set_architect

ENV ARCHITECT_ENDPOINT ""

COPY velocity/dist /usr/share/nginx/html
COPY docker/vhost.conf /etc/nginx/conf.d/default.conf