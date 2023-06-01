FROM alpine:3.14.1

RUN apk add --no-cache ca-certificates

ADD ./logging-operator /logging-operator

ENTRYPOINT ["/logging-operator"]
