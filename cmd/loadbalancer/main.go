package main

import (
	"context"
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

func main() {
	algorithm := flag.String("algorithm", "round-robin", "Load balancing algorithm: round-robin, least-connections, or weighted-response-time")
	listenAddr := flag.String("listen", ":8080", "Address to listen on")
	healthCheckInterval := flag.Duration("health-check-interval", 10*time.Second, "Interval between health checks")
	flag.Parse()

	serverURLs := []string{
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}

	servers := make([]*domain.Server, len(serverURLs))
	for i, urlStr := range serverURLs {
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
		Addr:    *listenAddr,
		Handler: handler,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go useCase.StartHealthCheck(ctx, *healthCheckInterval)

	go func() {
		log.Printf("Starting load balancer on %s using %s algorithm", *listenAddr, *algorithm)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
