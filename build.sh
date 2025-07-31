#!/bin/sh
set -e
mkdir -p bin
go build -o bin/web-service-go
echo "Built executable at bin/web-service-go"