```markdown
# Load Balancer Architecture

## Overview

This load balancer is designed using clean architecture principles, separating concerns into distinct layers:

1. Domain Layer
2. Use Cases Layer
3. Interfaces Layer
4. Infrastructure Layer

## Components

### Domain Layer

- Defines core business logic and entities (Server, LoadBalancer interface)

### Use Cases Layer

- Implements application-specific business rules
- Orchestrates the flow of data to and from entities

### Interfaces Layer

- Handles HTTP requests and responses
- Converts data between the format most convenient for entities and use cases

### Infrastructure Layer

- Implements concrete load balancing algorithms (Round Robin, Least Connections, Weighted Response Time)

## Flow

1. Incoming request handled by HTTP handler
2. Handler uses LoadBalancerUseCase to get next server
3. Request forwarded to selected server
4. Response from backend server returned to client

## Metrics and Monitoring

- Prometheus integration for collecting metrics
- Health checks performed periodically on backend servers
```
