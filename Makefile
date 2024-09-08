.PHONY: build test run clean docker-build docker-run

build:
    @echo "Building..."
    @go build -o bin/loadbalancer ./cmd/loadbalancer

test:
    @echo "Running tests..."
    @go test ./...

run: build
    @echo "Running..."
    @./bin/loadbalancer

clean:
    @echo "Cleaning..."
    @rm -rf bin/*

docker-build:
    @echo "Building Docker image..."
    @docker build -t loadbalancer:latest -f deployments/Dockerfile .

docker-run: docker-build
    @echo "Running Docker container..."
    @docker run -p 8080:8080 -p 9090:9090 loadbalancer:latest