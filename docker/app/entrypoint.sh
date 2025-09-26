#!/bin/sh
set -e

task setup
task -a

exec go run ./cmd/http
