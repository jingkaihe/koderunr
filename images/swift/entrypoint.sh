#!/bin/bash
set -e

source_code=$1
fname=$2.swift

touch $fname
echo "$source_code" > $fname
swiftc $fname -o main

./main
