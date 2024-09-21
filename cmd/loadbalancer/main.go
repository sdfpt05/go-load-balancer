package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/sdfpt05/go_load_balancer/v2/internal/domain"
	"github.com/sdfpt05/go_load_balancer/v2/internal/infrastructure/loadbalancers"
	"github.com/sdfpt05/go_load_balancer/v2/internal/interfaces"
	"github.com/sdfpt05/go_load_balancer/v2/internal/usecases"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

type Config struct {
	ListenAddr string   `json:"listen_addr"`
	Servers    []string `json:"servers"`
	TLSCert    string   `json:"tls_cert"`
	TLSKey     string   `json:"tls_key"`
}

func main() {
	configFile := flag.String("config", "config/config.json", "Path to configuration file")
	algorithm := flag.String("algorithm", "round-robin", "Load balancing algorithm: round-robin, least-connections, or weighted-response-time")
	healthCheckInterval := flag.Duration("health-check-interval", 10*time.Second, "Interval between health checks")
	flag.Parse()

	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	servers := make([]*domain.Server, len(config.Servers))
	for i, urlStr := range config.Servers {
		u, err := url.Parse(urlStr)
		if err != nil {
			log.Fatalf("Invalid server URL: %v", err)
		}
		servers[i] = &domain.Server{URL: u, Active: atomic.Bool{}}
		servers[i].Active.Store(true)
	}

	var lb domain.LoadBalancer
	switch *algorithm {
	case "round-robin":
		lb = loadbalancers.NewRoundRobin(servers)
	case "least-connections":
		lb = loadbalancers.NewLeastConnections(servers)
	case "weighted-response-time":
		lb = loadbalancers.NewWeightedResponseTime(servers)
	default:
		log.Fatalf("Unknown algorithm: %s", *algorithm)
	}

	useCase := usecases.NewLoadBalancerUseCase(lb)
	handler := interfaces.NewHTTPHandler(useCase)

	srv := &http.Server{
		Addr:    config.ListenAddr,
		Handler: handler,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go useCase.StartHealthCheck(ctx, *healthCheckInterval)

	go func() {
		log.Printf("Starting load balancer on %s using %s algorithm", config.ListenAddr, *algorithm)
		var err error
		if config.TLSCert != "" && config.TLSKey != "" {
			err = srv.ListenAndServeTLS(config.TLSCert, config.TLSKey)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen and serve error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

func loadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
