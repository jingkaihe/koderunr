#!/bin/bash
set -e

/bin/bash

source_code=$1
fname=$2.go

touch $fname
echo "$source_code" > $fname
cc $fname
./a.out
