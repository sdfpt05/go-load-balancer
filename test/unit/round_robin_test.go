package unit

import (
	"context"
	"net/url"
	"testing"

	"github.com/sdfpt05/go_load_balancer/v2/internal/domain"
	"github.com/sdfpt05/go_load_balancer/v2/internal/infrastructure/loadbalancers"
)

func TestRoundRobin(t *testing.T) {
	servers := []*domain.Server{
		{URL: mustParseURL("http://server1.com")},
		{URL: mustParseURL("http://server2.com")},
		{URL: mustParseURL("http://server3.com")},
	}

	rr := loadbalancers.NewRoundRobin(servers)

	for i := 0; i < 6; i++ {
		server, err := rr.NextServer(context.Background())
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		expectedURL := servers[i%3].URL
		if server.URL != expectedURL {
			t.Errorf("Expected server %s, got %s", expectedURL, server.URL)
		}
	}
}

func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}
