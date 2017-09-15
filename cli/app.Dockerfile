FROM scratch

ADD velocity/dist/velocity /bin/velocity

ENTRYPOINT ["/bin/velocity"]
