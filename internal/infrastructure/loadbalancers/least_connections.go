package loadbalancers

import (
	"context"
	"github.com/sdfpt05/go_load_balancer/v2/internal/domain"
	"sort"
)

type LeastConnections struct {
	BaseLoadBalancer
}

func NewLeastConnections(servers []*domain.Server) *LeastConnections {
	return &LeastConnections{BaseLoadBalancer: BaseLoadBalancer{servers: servers}}
}

func (lc *LeastConnections) NextServer(ctx context.Context) (*domain.Server, error) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	if len(lc.servers) == 0 {
		return nil, ErrNoServersAvailable
	}

	activeServers := make([]*domain.Server, 0, len(lc.servers))
	for _, server := range lc.servers {
		if server.Active.Load() {
			activeServers = append(activeServers, server)
		}
	}

	if len(activeServers) == 0 {
		return nil, ErrNoServersAvailable
	}

	sort.Slice(activeServers, func(i, j int) bool {
		return activeServers[i].Connections < activeServers[j].Connections
	})

	return activeServers[0], nil
}
