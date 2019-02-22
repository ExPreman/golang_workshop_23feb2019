#!/bin/sh

set -eu
cd "$(dirname "$0")/.."

mkdir -p output

export GO111MODULE=on
export GOBIN="$PWD/output"
export CGO_ENABLED=0

go install ./cmd/...
