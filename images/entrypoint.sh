#!/bin/bash
set -e

case "$1" in
  ".go" )
    touch runner.go
    echo "$2" > runner.go
    go run runner.go ;;
  ".c" )
    touch runner.c
    echo "$2" > runner.c
    cc runner.c
    ./a.out ;;
  ".rb" )
    touch runner.rb
    echo "$2" > runner.rb
    ruby runner.rb ;;
  ".py" )
    touch runner.py
    echo "$2" > runner.py
    python runner.py ;;
  * )
    exec "$@"
esac
