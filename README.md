# Go Load Balancer

A high-performance, feature-rich TCP load balancer written in Go.

## Features

- Multiple load balancing algorithms: Round Robin, Least Connections, Weighted Response Time
- Periodic health checks of backend servers
- Rate limiting
- Circuit breaker pattern for improved fault tolerance
- Dynamic server management (add/remove servers at runtime)
- Optional TLS support
- Prometheus metrics for monitoring
- Configurable via YAML file

## Requirements

- Go 1.16 or later
- Docker (for containerized deployment)

## Configuration

Create a `config.yaml` file in the `config` directory:

```yaml
server:
  listen_addr: "127.0.0.1:8080"
  read_timeout: 5s
  write_timeout: 10s
  idle_timeout: 120s

load_balancer:
  algorithm: "round-robin"
  health_check_interval: 10s

backend_servers:
  - "http://localhost:8081"
  - "http://localhost:8082"
  - "http://localhost:8083"

tls:
  enabled: false
  cert_file: ""
  key_file: ""

logging:
  level: "info"
  format: "json"

metrics:
  enabled: true
  port: 9090
```
## Building and Running

Use the provided Makefile:

- `make build`: Build the load balancer
- `make run`: Run the load balancer
- `make test`: Run all tests
- `make clean`: Remove built binaries
- `make docker-build`: Build Docker image
- `make docker-run`: Run the load balancer in a Docker container

## Usage

### Starting the Load Balancer

```bash
make run
```
### Adding a Server

```bash
curl -X POST -H "Content-Type: application/json" -d '{"url":"http://newserver:8080"}' http://localhost:8080/servers
```
### Removing a Server
```bash
curl -X DELETE http://localhost:8080/servers/http%3A%2F%2Fnewserver%3A8080
```
### Checking Load Balancer Health
```bash
curl http://localhost:8080/health
```
## Metrics
Prometheus metrics are available at http://localhost:9090/metrics when enabled in the configuration.

## Testing
Run the test suite:
``` bash
make test
```
## Deployment
For containerized deployment:
``` bash
make docker-run
```
## License

This project is licensed under the MIT License.
