package loadbalancers

import (
	"context"
	"myapp/domain"
	"net/http"
	"sync"
	"time"
)

type BaseLoadBalancer struct {
	servers []*domain.Server
	mu      sync.RWMutex
}

func (b *BaseLoadBalancer) HealthCheck(ctx context.Context) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, server := range b.servers {
		go func(s *domain.Server) {
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, "GET", s.URL.String()+"/health", nil)
			if err != nil {
				s.Active.Store(false)
				return
			}

			start := time.Now()
			resp, err := http.DefaultClient.Do(req)
			duration := time.Since(start)

			if err != nil || resp.StatusCode != http.StatusOK {
				s.Active.Store(false)
			} else {
				s.Active.Store(true)
				s.ResponseTime = duration
			}
			s.LastChecked = time.Now()
		}(server)
	}
}
