FROM gsoci.azurecr.io/giantswarm/alpine:3.20.3-giantswarm

WORKDIR /

ADD logging-operator logging-operator

USER 65532:65532

ENTRYPOINT ["/logging-operator"]
