#!/bin/bash
set -e

source_code=$1
fname=$2.py

if [[ -n "$3" ]]; then
  py_version=$3
  py_versions=($(cd ~/.pyenv/versions && ls -d */))
  for version in $py_versions
  do
    if [[ $version == *"$py_version"* ]]
    then
      py_version=$version
      break
    fi
  done
  pyenv global $py_version
fi

touch $fname
echo "$source_code" > $fname
python $fname
