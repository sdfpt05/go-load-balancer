#!/bin/sh
set -e

echo "Deploying load balancer..."

docker-compose -f deployments/docker-compose.yml up -d

echo "Deployment complete."