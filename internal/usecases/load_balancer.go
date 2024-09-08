package usecases

import (
	"context"
	"github.com/sdfpt05/go_load_balancer/v2/internal/domain"
	"time"
)

type LoadBalancerUseCase struct {
	lb domain.LoadBalancer
}

func NewLoadBalancerUseCase(lb domain.LoadBalancer) *LoadBalancerUseCase {
	return &LoadBalancerUseCase{lb: lb}
}

func (uc *LoadBalancerUseCase) GetNextServer(ctx context.Context) (*domain.Server, error) {
	return uc.lb.NextServer(ctx)
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
