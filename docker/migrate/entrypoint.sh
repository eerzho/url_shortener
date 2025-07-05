#!/bin/sh
if [ $# -eq 0 ]; then
    exec migrate -path /migrations -database "$DB_URL" up
else
    exec migrate -path /migrations -database "$DB_URL" "$@"
fi
