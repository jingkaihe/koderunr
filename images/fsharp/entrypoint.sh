#!/bin/sh
set -e

source_code=$1
fname=Program.fs

touch $fname

echo "$source_code" > $fname

dotnet run $fname
