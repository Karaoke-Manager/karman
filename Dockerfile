ARG TARGETPLATFORM=linux/amd64
ARG TARGETOS=linux
ARG TARGETARCH=amd64
FROM --platform=$TARGETPLATFORM alpine:3.18

ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache ca-certificates && \
    adduser --disabled-password --no-create-home --uid 1008 karman
COPY --chmod=755 build/${TARGETOS}/${TARGETARCH}/karman /usr/local/bin/karman
USER karman:karman

EXPOSE 8080
VOLUME /usr/local/share/karman
ENTRYPOINT ["karman", "server"]
