#!/bin/bash

set -euo pipefail

cd updater
wire
go generate
go test -count=1 ./...