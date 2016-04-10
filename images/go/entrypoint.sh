#!/bin/bash
set -e

source_code=$1
fname=$2.go

touch $fname
echo "$source_code" > $fname
go run $fname
