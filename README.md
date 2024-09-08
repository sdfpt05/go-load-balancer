# load-balancer

A simple TCP load balancer written in Go I built to test one of my many pet projects.

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

## Building and Running

Use the provided Makefile:

- **`make build`**: Build the load balancer and test server.
- **`make run`**: Run the load balancer.
- **`make test_servers`**: Run three test servers.
- **`make clean`**: Remove built binaries.

## Testing

Start the test servers:

```bash
make test_servers
```

In another terminal, start the load balancer:

```bash
make run
```

Send requests to the load balancer:

```bash
curl http://localhost:8080
```

## License

This project is licensed under the MIT License.
