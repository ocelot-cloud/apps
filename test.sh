#!/bin/bash

set -euo pipefail

cd updater
wire
go generate

echo "Running unit tests..."
go test ./...

echo "Running integration tests..."
go test -tags=integration ./...