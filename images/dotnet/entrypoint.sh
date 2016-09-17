#!/bin/sh
set -e

source_code=$1
fname=runner.cs

touch $fname

echo "$source_code" > $fname

dotnet run $fname
