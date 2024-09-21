package loadbalancers

import (
	"context"
	"github.com/sdfpt05/go_load_balancer/v2/internal/domain"
	"sort"
)

type WeightedResponseTime struct {
	BaseLoadBalancer
}

func NewWeightedResponseTime(servers []*domain.Server) *WeightedResponseTime {
	return &WeightedResponseTime{BaseLoadBalancer: BaseLoadBalancer{servers: servers}}
}

func (wrt *WeightedResponseTime) NextServer(ctx context.Context) (*domain.Server, error) {
	wrt.mu.RLock()
	defer wrt.mu.RUnlock()

	if len(wrt.servers) == 0 {
		return nil, ErrNoServersAvailable
	}

	activeServers := make([]*domain.Server, 0, len(wrt.servers))
	for _, server := range wrt.servers {
		if server.Active.Load() {
			activeServers = append(activeServers, server)
		}
	}

	if len(activeServers) == 0 {
		return nil, ErrNoServersAvailable
	}

	sort.Slice(activeServers, func(i, j int) bool {
		return activeServers[i].ResponseTime < activeServers[j].ResponseTime
	})

	return activeServers[0], nil
}
