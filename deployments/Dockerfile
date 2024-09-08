FROM golang:1.19-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o loadbalancer ./cmd/loadbalancer

FROM alpine:3.14

WORKDIR /app

COPY --from=builder /app/loadbalancer .
COPY configs/config.yaml /app/configs/

EXPOSE 8080 9090

CMD ["./loadbalancer"]