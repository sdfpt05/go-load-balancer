# load-balancer

This project implements a simple TCP load balancer with health checks, written in Go.

## Features

- TCP load balancing with least-connections algorithm
- Periodic health checks of backend servers
- Optional TLS support
- Configurable via JSON file

## Requirements

- Go 1.16 or later

## Configuration

Create a `config.json` file in the project root:

```json
{
  "listen_addr": "127.0.0.1:8080",
  "servers": ["127.0.0.1:8081", "127.0.0.1:8082", "127.0.0.1:8083"],
  "tls_cert": "",
  "tls_key": ""
}
```
