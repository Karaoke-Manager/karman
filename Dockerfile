ARG TARGETPLATFORM=linux/amd64
ARG TARGETOS=linux
ARG TARGETARCH=amd64
FROM --platform=$TARGETPLATFORM gcr.io/distroless/static

ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

COPY --chmod=755 build/${TARGETOS}/${TARGETARCH}/karman /karman
USER nonroot:nonroot

EXPOSE 8080
ENTRYPOINT ["/karman", "server"]
