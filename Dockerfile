FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY rss-email .
COPY config.yaml .
USER nonroot:nonroot
CMD ["./rss-email"]
