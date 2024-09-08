#!/bin/bash

# Create directories
mkdir -p internal/domain
mkdir -p internal/usecases
mkdir -p internal/interfaces
mkdir -p internal/infrastructure/loadbalancers
mkdir -p pkg/metrics
mkdir -p configs
mkdir -p scripts
mkdir -p deployments
mkdir -p api
mkdir -p test/unit
mkdir -p test/integration
mkdir -p docs

# Create files with placeholder text
touch internal/domain/server.go
touch internal/domain/load_balancer.go

touch internal/usecases/load_balancer.go

touch internal/interfaces/http_handler.go

touch internal/infrastructure/loadbalancers/base.go
touch internal/infrastructure/loadbalancers/round_robin.go
touch internal/infrastructure/loadbalancers/least_connections.go
touch internal/infrastructure/loadbalancers/weighted_response_time.go

mkdir pkg/metrics/prometheus.go

touch configs/config.yaml

touch scripts/build.sh
touch scripts/deploy.sh

touch deployments/docker-compose.yml
touch deployments/Dockerfile

touch api/openapi.yaml

touch test/unit/.gitkeep
touch test/integration/.gitkeep

touch docs/architecture.md
touch docs/api.md

touch go.mod
touch go.sum
touch .gitignore
touch README.md
touch Makefile

echo "Project structure created successfully!"
