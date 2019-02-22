#!/bin/sh

set -eu
cd "$(dirname "$0")/.."

mkdir -p output

export GO111MODULE=on

go test -race ./...
