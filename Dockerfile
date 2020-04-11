FROM alpine

COPY hooker /bin/hooker

ENTRYPOINT /bin/hooker
