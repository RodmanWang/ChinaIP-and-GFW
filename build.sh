#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"
mkdir -p dist

go build -trimpath -ldflags='-s -w' -o dist/ipip2mmdb .
go build -trimpath -ldflags='-s -w' -o dist/verify_ip ./verify
