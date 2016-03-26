#!/bin/bash
set -e

case "$1" in
  ".go" )
    touch runner.go
    echo "$2" > runner.go
    go run runner.go ;;
  * )
    exec "$@"
esac
