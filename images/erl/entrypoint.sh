#!/bin/bash
set -e

source_code=$1
fname=$2.ex

touch $fname
echo "$source_code" > $fname
elixir $fname
