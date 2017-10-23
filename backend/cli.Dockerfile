FROM scratch

ADD dist/velocity_cli /bin/velocity

ENTRYPOINT ["/bin/velocity"]
