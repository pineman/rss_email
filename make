#!/usr/bin/env bash

set -e

deploy() {
  fly deploy
}

eval "${@:-deploy}"
