FROM gcr.io/distroless/static

LABEL maintainer="codello"

USER nonroot:nonroot

COPY build/karman build/migrate /usr/local/bin/

ENTRYPOINT ["karman"]