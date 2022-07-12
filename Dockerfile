FROM gcr.io/distroless/static

COPY depserver-linux-amd64 /depserver
COPY depsyncer-linux-amd64 /depsyncer

EXPOSE 9001

VOLUME ["/db"]

ENTRYPOINT ["/depserver"]
