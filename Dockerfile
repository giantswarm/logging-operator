# Use distroless as minimal base image to package the logging-operator binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
ADD bin/logging-operator .
USER 65532:65532

ENTRYPOINT ["/logging-operator"]
