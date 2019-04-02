FROM alpine:3.4

WORKDIR /knot

RUN apk update \
    && apk add --no-cache ca-certificates bash \
    && rm -rf /var/cache/apk/* \
    && mkdir /knot/files

COPY artifacts/knot-linux.tgz .
RUN tar xzvf ./knot-linux.tgz

CMD ["./knot"]

