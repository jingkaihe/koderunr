#!/bin/sh
set -e

source_code=$1
fname=$2.c

touch $fname
echo "$source_code" > $fname
cc $fname
./a.out
