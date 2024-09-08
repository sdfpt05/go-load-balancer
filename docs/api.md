```markdown
# Load Balancer API

## Endpoints

### GET /

- Description: Forwards the request to a backend server based on the chosen load balancing algorithm
- Response: The response from the backend server

### GET /health

- Description: Returns the health status of the load balancer
- Response: 200 OK if the load balancer is healthy

### GET /metrics

- Description: Returns Prometheus metrics
- Response: Prometheus formatted metrics data
```
