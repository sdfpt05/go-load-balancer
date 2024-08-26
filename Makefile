
.PHONY: build run test_servers clean

build:
	go build -o bin/load_balancer cmd/lb/main.go
	go build -o bin/test_server cmd/be/test_server.go

run: build
	./bin/load_balancer

test_servers: build
	./bin/test_server -port 8081 & \
	./bin/test_server -port 8082 & \
	./bin/test_server -port 8083 &

clean:
	rm -f bin/load_balancer bin/test_server