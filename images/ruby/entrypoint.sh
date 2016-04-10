#!/bin/bash
set -e

source_code=$1
fname=$2.rb

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

touch $fname
echo "$source_code" > $fname
ruby $fname
