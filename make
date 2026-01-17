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

eval "${@:-deploy}"
