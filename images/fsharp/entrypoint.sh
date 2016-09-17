#!/bin/sh
set -e

source_code=$1
fname=Runner.fs

touch $fname

echo "$source_code" > $fname

dotnet publish > /dev/null

dotnet bin/Debug/netcoreapp1.0/publish/fsharp.dll
