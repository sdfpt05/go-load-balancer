BINARY_NAME=load_balancer
DOCKER_IMAGE_NAME=go-load-balancer

.PHONY: all build test clean run docker-build docker-run

all: test build

build:
	go build -o ${BINARY_NAME} cmd/loadbalancer/main.go

test:
	go test -v ./...

clean:
	go clean
	rm -f ${BINARY_NAME}

run: build
	./${BINARY_NAME}

docker-build:
	docker build -t ${DOCKER_IMAGE_NAME} .

docker-run: docker-build
	docker run -p 8080:8080 -p 9090:9090 ${DOCKER_IMAGE_NAME}

lint:
	golangci-lint run

benchmark:
	go test -bench=. ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

deps:
	go mod tidy
	go mod verify