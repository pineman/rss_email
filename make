#!/usr/bin/env bash

set -euxo pipefail

run() {
  go run .
}

build() {
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o rss-email .
}

deploy() {
  build
  fly deploy --local-only
}

logs() {
  fly ssh console -C "sh -c 'tail -n +1 -f /app/data/logs/rss_email_*.log'"
}

query() {
  local q="${*//\"/\\\"}"
  fly ssh console -C "sqlite3 /app/data/rss_email.db \"$q\""
}

"${@:-deploy}"
