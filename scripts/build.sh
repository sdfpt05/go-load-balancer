#!/bin/sh
set -e

echo "Building load balancer..."
go build -o bin/loadbalancer ./cmd/loadbalancer

echo "Build complete."