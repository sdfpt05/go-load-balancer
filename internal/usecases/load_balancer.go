package usecases

import (
	"context"
	"time"

	"github.com/sdfpt05/go_load_balancer/v2/internal/domain"
	"github.com/sdfpt05/go_load_balancer/v2/internal/infrastructure/circuitbreaker"
)

type LoadBalancerUseCase struct {
	lb            domain.LoadBalancer
	circuitBreaker *circuitbreaker.CircuitBreaker
}

func NewLoadBalancerUseCase(lb domain.LoadBalancer, cb *circuitbreaker.CircuitBreaker) *LoadBalancerUseCase {
	return &LoadBalancerUseCase{
		lb:            lb,
		circuitBreaker: cb,
	}
}

func (uc *LoadBalancerUseCase) GetNextServer(ctx context.Context) (*domain.Server, error) {
	var server *domain.Server
	err := uc.circuitBreaker.Execute(func() error {
		var err error
		server, err = uc.lb.NextServer(ctx)
		return err
	})
	return server, err
}

func (uc *LoadBalancerUseCase) UpdateServerStatus(server *domain.Server) {
	uc.lb.UpdateServer(server)
}

func (uc *LoadBalancerUseCase) StartHealthCheck(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			uc.lb.HealthCheck(ctx)
		}
	}
}

func (uc *LoadBalancerUseCase) AddServer(server *domain.Server) error {
	return uc.lb.AddServer(server)
}

func (uc *LoadBalancerUseCase) RemoveServer(url string) error {
	return uc.lb.RemoveServer(url)
}

func (uc *LoadBalancerUseCase) GetServers() []*domain.Server {
	return uc.lb.GetServers()
}
