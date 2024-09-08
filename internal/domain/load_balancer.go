package domain

import "context"

type LoadBalancer interface {
	NextServer(ctx context.Context) (*Server, error)
	UpdateServer(server *Server)
	HealthCheck(ctx context.Context)
}
