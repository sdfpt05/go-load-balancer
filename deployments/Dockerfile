# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o load_balancer cmd/loadbalancer/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/load_balancer .
COPY config/config.yaml .

EXPOSE 8080 9090

CMD ["./load_balancer", "-config", "config.yaml"]