FROM alpine:3.20
COPY mctl /usr/bin/mctl
ENTRYPOINT ["/usr/bin/mctl"]