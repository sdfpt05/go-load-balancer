package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sdfpt05/go_load_balancer/v2/internal/domain"
	"github.com/sdfpt05/go_load_balancer/v2/internal/infrastructure/circuitbreaker"
	"github.com/sdfpt05/go_load_balancer/v2/internal/infrastructure/loadbalancers"
	"github.com/sdfpt05/go_load_balancer/v2/internal/interfaces"
	"github.com/sdfpt05/go_load_balancer/v2/internal/usecases"
)

func TestLoadBalancerIntegration(t *testing.T) {
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Backend 1"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Backend 2"))
	}))
	defer backend2.Close()

	servers := []*domain.Server{
		{URL: mustParseURL(backend1.URL)},
		{URL: mustParseURL(backend2.URL)},
	}

	lb := loadbalancers.NewRoundRobin(servers)
	cb := circuitbreaker.NewCircuitBreaker(5, 10*time.Second)
	useCase := usecases.NewLoadBalancerUseCase(lb, cb)

	logger, _ := zap.NewDevelopment()
	handler := interfaces.NewHTTPHandler(useCase, logger)

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}
}

func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}
