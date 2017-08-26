FROM erlang:slim

# HTTP
EXPOSE 80
# EPMD
EXPOSE 4369
EXPOSE 9100-9155
EXPOSE 45892/udp

ENV PORT=80
ENV MIX_ENV=prod
ENV REPLACE_OS_VARS=true
ENV SHELL=/bin/bash

WORKDIR /opt/app

ADD velocity/_build/prod/rel/velocity/releases/0.0.1/velocity.tar.gz ./

# Add wait-for-it
ADD https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh /bin/wait-for-it.sh
RUN chmod +x /bin/wait-for-it.sh

# Add entrypoint
COPY velocity/rel/entrypoint.sh /opt/app/entrypoint.sh
RUN chmod +x /opt/app/entrypoint.sh

ENTRYPOINT ["/opt/app/entrypoint.sh"]

CMD foreground
