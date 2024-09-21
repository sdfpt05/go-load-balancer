#!/bin/bash
set -e

echo "Starting deployment process..."

# Pre-deployment checks
echo "Running pre-deployment checks..."
if ! docker info > /dev/null 2>&1; then
    echo "Error: Docker is not running or not accessible"
    exit 1
fi

if [ ! -f "deployments/docker-compose.yml" ]; then
    echo "Error: docker-compose.yml file not found"
    exit 1
fi

# Deploy
echo "Deploying load balancer..."
docker-compose -f deployments/docker-compose.yml up -d

# Post-deployment verification
echo "Running post-deployment checks..."
if ! curl -sSf http://localhost:8080/health > /dev/null 2>&1; then
    echo "Error: Load balancer health check failed"
    exit 1
fi

echo "Deployment completed successfully."