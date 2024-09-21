package domain

import (
	"fmt"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
)

type Server struct {
	URL             *url.URL
	Active          atomic.Bool
	Connections     int64
	LastChecked     time.Time
	ResponseTime    time.Duration
	HealthCheckPath string
	Weight          int
	FailureCount    int
}

func NewServer(urlStr string) (*Server, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	server := &Server{
		URL:             u,
		HealthCheckPath: "/health", // Default health check path
		Weight:          1,         // Default weight
	}
	server.Active.Store(true)

	return server, nil
}

func (s *Server) HealthCheck() error {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(s.URL.String() + s.HealthCheckPath)
	if err != nil {
		s.FailureCount++
		return err
	}
	defer resp.Body.Close()

	s.ResponseTime = time.Since(s.LastChecked)
	s.LastChecked = time.Now()

	if resp.StatusCode != http.StatusOK {
		s.FailureCount++
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	s.FailureCount = 0
	s.Active.Store(true)
	return nil
}
