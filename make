#!/usr/bin/env bash

set -e

run() {
  go run .
}

deploy() {
  fly deploy --local-only
}

eval "${@:-deploy}"
