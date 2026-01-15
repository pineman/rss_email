FROM golang:1.25.5-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o rss-email .

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /build/rss-email .
COPY --from=builder /build/config.yaml .
RUN mkdir -p /app/data
RUN adduser -D -u 1000 appuser && \
    chown -R appuser:appuser /app
USER appuser
CMD ["./rss-email"]
