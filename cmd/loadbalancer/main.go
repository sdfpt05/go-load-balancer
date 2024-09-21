package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sdfpt05/go_load_balancer/v2/internal/config"
	"github.com/sdfpt05/go_load_balancer/v2/internal/domain"
	"github.com/sdfpt05/go_load_balancer/v2/internal/infrastructure/circuitbreaker"
	"github.com/sdfpt05/go_load_balancer/v2/internal/infrastructure/loadbalancers"
	"github.com/sdfpt05/go_load_balancer/v2/internal/interfaces"
	"github.com/sdfpt05/go_load_balancer/v2/internal/middleware"
	"github.com/sdfpt05/go_load_balancer/v2/internal/usecases"
	"github.com/sdfpt05/go_load_balancer/v2/pkg/metrics"
	"go.uber.org/zap"
)

func main() {
	configFile := flag.String("config", "config/config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup logger
	logger, err := setupLogger(cfg.Logging)
	if err != nil {
		log.Fatalf("Failed to setup logger: %v", err)
	}
	defer logger.Sync()

	// Setup metrics
	if cfg.Metrics.Enabled {
		metrics.Setup(cfg.Metrics.Port)
	}

	// Initialize servers
	servers, err := initializeServers(cfg.BackendServers)
	if err != nil {
		logger.Fatal("Failed to initialize servers", zap.Error(err))
	}

	// Initialize load balancer
	lb, err := initializeLoadBalancer(cfg.LoadBalancer.Algorithm, servers)
	if err != nil {
		logger.Fatal("Failed to initialize load balancer", zap.Error(err))
	}

	// Initialize circuit breaker
	cb := circuitbreaker.NewCircuitBreaker(5, 10*time.Second)

	useCase := usecases.NewLoadBalancerUseCase(lb, cb)

	// Initialize rate limiter
	rl := middleware.NewRateLimiter(100, 10) // 100 requests per second, burst of 10

	handler := interfaces.NewHTTPHandler(useCase, logger)

	// Setup server
	srv := &http.Server{
		Addr:         cfg.Server.ListenAddr,
		Handler:      rl.RateLimit(handler),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start health check
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go useCase.StartHealthCheck(ctx, cfg.LoadBalancer.HealthCheckInterval)

	// Start server
	go func() {
		logger.Info("Starting load balancer", zap.String("address", cfg.Server.ListenAddr))
		if cfg.TLS.Enabled {
			if err := srv.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile); err != nil && err != http.ErrServerClosed {
				logger.Fatal("Failed to start server", zap.Error(err))
			}
		} else {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatal("Failed to start server", zap.Error(err))
			}
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exiting")
}

func setupLogger(cfg config.LoggingConfig) (*zap.Logger, error) {
	var zapConfig zap.Config
	if cfg.Format == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	zapConfig.Level = zap.NewAtomicLevelAt(logLevelFromString(cfg.Level))
	return zapConfig.Build()
}

func logLevelFromString(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func initializeServers(urls []string) ([]*domain.Server, error) {
	servers := make([]*domain.Server, len(urls))
	for i, urlStr := range urls {
		u, err := url.Parse(urlStr)
		if err != nil {
			return nil, fmt.Errorf("invalid server URL %s: %w", urlStr, err)
		}
		servers[i] = &domain.Server{URL: u, Active: atomic.Bool{}}
		servers[i].Active.Store(true)
	}
	return servers, nil
}

func initializeLoadBalancer(algorithm string, servers []*domain.Server) (domain.LoadBalancer, error) {
	switch algorithm {
	case "round-robin":
		return loadbalancers.NewRoundRobin(servers), nil
	case "least-connections":
		return loadbalancers.NewLeastConnections(servers), nil
	case "weighted-response-time":
		return loadbalancers.NewWeightedResponseTime(servers), nil
	default:
		return nil, fmt.Errorf("unknown algorithm: %s", algorithm)
	}
}
