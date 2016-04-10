#!/bin/bash
set -e

ext=$1
source_code=$2
fname=$3

case "$ext" in
  "go" )
    touch $uuid.go
    echo "$source_code" > $uuid.go
    go run $uuid.go ;;
  "c" )
    touch $uuid.c
    echo "$source_code" > $uuid.c
    cc $uuid.c
    ./a.out ;;
  "ruby" )
    if [[ -n "$4" ]]; then
      rb_version=$4
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

    touch $uuid.rb
    echo "$source_code" > $uuid.rb
    ruby $uuid.rb ;;
  "python" )
    touch $uuid.py
    echo "$source_code" > $uuid.py
    python $uuid.py ;;
  "elixir" )
    touch $uuid.ex
    echo "$source_code" > $uuid.ex
    elixir $uuid.ex ;;
esac
