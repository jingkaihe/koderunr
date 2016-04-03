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
    if [[ -n "$3" ]]; then
      rb_version=$3
      rb_versions=($(cd ~/.rbenv/versions && ls -d */))
      for version in $rb_versions
      do
        if [[ $version == *"$rb_version"* ]]
        then
          rb_version=$version
          break
        fi
      done
      rbenv global $rb_version
    fi

    touch runner.rb
    echo "$2" > runner.rb
    ruby runner.rb ;;
  ".py" )
    touch runner.py
    echo "$2" > runner.py
    python runner.py ;;
esac
