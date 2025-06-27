#!/bin/bash

set -euo pipefail

cd updater
go generate
wire
go test -count=1 ./...