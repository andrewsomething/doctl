#!/usr/bin/env bash

set -euo pipefail

base="-X github.com/digitalocean/doctl."
build="$("$DIR"/scripts/version.sh -c)"
ldflags="${base}Build=${build}"

version="$("$DIR"/scripts/version.sh -s)"
major="$(echo "$version" | cut -d . -f1)"
ldflags="${ldflags} ${base}Major=${major}"

minor="$(echo "$version" | cut -d . -f2)"
ldflags="${ldflags} ${base}Minor=${minor}"

patch="$(echo "$version" | cut -d . -f3)"
ldflags="${ldflags} ${base}Patch=${patch}"

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../"
OUT_D=${OUT_D:-${DIR}/builds}
mkdir -p "$OUT_D"

(
  export GOOS=${GOOS:-linux}
  export GOARCH=${GOARCH:-amd64}
  export GOFLAGS=-mod=vendor
  cd cmd/doctl && go build -ldflags "$ldflags" -o "${OUT_D}/doctl_${GOOS}_${GOARCH}"
)
