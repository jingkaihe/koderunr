#!/bin/sh
set -e

source_code=$1
fname=$2.rb

touch $fname
echo "$source_code" > $fname
ruby $fname
