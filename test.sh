#!/bin/bash

set -euo pipefail

cd updater
go generate
go test -count=1 ./...