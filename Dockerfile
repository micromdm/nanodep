FROM gcr.io/distroless/static

ARG TARGETOS TARGETARCH

COPY depserver-$TARGETOS-$TARGETARCH /app/depserver
COPY depsyncer-$TARGETOS-$TARGETARCH /app/depsyncer

EXPOSE 9001

VOLUME ["/app/db"]

WORKDIR /app

ENTRYPOINT ["/app/depserver"]
