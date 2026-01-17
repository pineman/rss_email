FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata sqlite tini
WORKDIR /app
COPY rss-email .
COPY config.yaml .
RUN mkdir -p /app/data
RUN adduser -D -u 1000 appuser && \
    chown -R appuser:appuser /app
USER appuser
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["./rss-email"]
