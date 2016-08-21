#!/bin/sh
set -e

source_code=$1
fname=$2.py

touch $fname
echo "$source_code" > $fname
python $fname
