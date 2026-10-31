#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
go build -ldflags "-H=windowsgui -s -w" -o umango.exe .
