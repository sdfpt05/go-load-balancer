package loadbalancers

import (
	"context"
	"myapp/domain"
	"sync/atomic"
)

type RoundRobin struct {
	BaseLoadBalancer
	current int64
}

func NewRoundRobin(servers []*domain.Server) *RoundRobin {
	return &RoundRobin{BaseLoadBalancer: BaseLoadBalancer{servers: servers}}
}

func (rr *RoundRobin) NextServer(ctx context.Context) (*domain.Server, error) {
	rr.mu.RLock()
	defer rr.mu.RUnlock()

	if len(rr.servers) == 0 {
		return nil, ErrNoServersAvailable
	}

	startIndex := int(atomic.AddInt64(&rr.current, 1) % int64(len(rr.servers)))
	for i := 0; i < len(rr.servers); i++ {
		index := (startIndex + i) % len(rr.servers)
		if rr.servers[index].Active.Load() {
			return rr.servers[index], nil
		}
	}

	return nil, ErrNoServersAvailable
}

func (rr *RoundRobin) UpdateServer(server *domain.Server) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	for i, s := range rr.servers {
		if s.URL.String() == server.URL.String() {
			rr.servers[i] = server
			break
		}
	}
}
