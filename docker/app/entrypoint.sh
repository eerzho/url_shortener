#!/bin/sh
set -e

task setup
task -a

exec go run ./cmd/app/main.go
