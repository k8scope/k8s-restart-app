FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY restart-app /restart-app
USER 65532:65532
ENTRYPOINT ["/restart-app"]
