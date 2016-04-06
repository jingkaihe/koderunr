#!/bin/bash
set -e

ext=$1
source_code=$2

case "$ext" in
  ".go" )
    touch runner.go
    echo "$source_code" > runner.go
    go run runner.go ;;
  ".c" )
    touch runner.c
    echo "$source_code" > runner.c
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
    echo "$source_code" > runner.rb
    ruby runner.rb ;;
  ".py" )
    touch runner.py
    echo "$source_code" > runner.py
    python runner.py ;;
  ".ex" )
    touch runner.ex
    echo "$source_code" > runner.ex
    elixir runner.ex ;;
esac
