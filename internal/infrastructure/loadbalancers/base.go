package loadbalancers

import (
	"context"
	"github.com/sdfpt05/go_load_balancer/v2/internal/domain"
	"sync"
	"errors"
)

type BaseLoadBalancer struct {
	servers []*domain.Server
	mu      sync.RWMutex
}

func (b *BaseLoadBalancer) UpdateServer(server *domain.Server) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i, s := range b.servers {
		if s.URL.String() == server.URL.String() {
			b.servers[i] = server
			break
		}
	}
}

func (b *BaseLoadBalancer) HealthCheck(ctx context.Context) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, server := range b.servers {
		go func(s *domain.Server) {
			if err := s.HealthCheck(); err != nil {
				s.Active.Store(false)
			} else {
				s.Active.Store(true)
			}
		}(server)
	}
}

func (b *BaseLoadBalancer) AddServer(server *domain.Server) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.servers = append(b.servers, server)
	return nil
}

func (b *BaseLoadBalancer) RemoveServer(url string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, s := range b.servers {
		if s.URL.String() == url {
			b.servers = append(b.servers[:i], b.servers[i+1:]...)
			return nil
		}
	}
	return errors.New("server not found")
}

func (b *BaseLoadBalancer) GetServers() []*domain.Server {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return append([]*domain.Server{}, b.servers...)
}