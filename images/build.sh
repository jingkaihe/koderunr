#!/bin/bash

set -e

platforms=(python ruby go erl c swift)

for platform in "${platforms[@]}"
do
  echo "building $platform image..."
  docker build -t koderunr-$platform ./$platform
  echo "$platform image is built successfully!"
done
