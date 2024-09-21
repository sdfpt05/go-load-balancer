package interfaces

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/sdfpt05/go_load_balancer/v2/internal/domain"
	"github.com/sdfpt05/go_load_balancer/v2/internal/usecases"
	"github.com/sdfpt05/go_load_balancer/v2/pkg/metrics"
	"go.uber.org/zap"
)

type HTTPHandler struct {
	loadBalancerUseCase *usecases.LoadBalancerUseCase
	logger              *zap.Logger
}

func NewHTTPHandler(uc *usecases.LoadBalancerUseCase, logger *zap.Logger) *HTTPHandler {
	return &HTTPHandler{loadBalancerUseCase: uc, logger: logger}
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logger := h.logger.With(zap.String("request_id", r.Header.Get("X-Request-ID")))

	logger.Info("Incoming request", zap.String("path", r.URL.Path))

	switch r.URL.Path {
	case "/health":
		h.handleHealth(w, r)
	case "/servers":
		switch r.Method {
		case http.MethodGet:
			h.handleGetServers(w, r)
		case http.MethodPost:
			h.handleAddServer(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		h.handleProxy(w, r)
	}

	duration := time.Since(startTime)
	logger.Info("Request completed", zap.Duration("duration", duration))
	metrics.RequestDuration.Observe(duration.Seconds())
}

func (h *HTTPHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *HTTPHandler) handleGetServers(w http.ResponseWriter, r *http.Request) {
	servers := h.loadBalancerUseCase.GetServers()
	json.NewEncoder(w).Encode(servers)
}

func (h *HTTPHandler) handleAddServer(w http.ResponseWriter, r *http.Request) {
	var serverInput struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&serverInput); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	server, err := domain.NewServer(serverInput.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.loadBalancerUseCase.AddServer(server); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *HTTPHandler) handleProxy(w http.ResponseWriter, r *http.Request) {
	server, err := h.loadBalancerUseCase.GetNextServer(r.Context())
	if err != nil {
		http.Error(w, "No server available", http.StatusServiceUnavailable)
		h.logger.Error("No server available", zap.Error(err))
		metrics.RequestsTotal.WithLabelValues("error").Inc()
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(server.URL)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		h.logger.Error("Proxy error", zap.Error(err))
		server.Active.Store(false)
		h.loadBalancerUseCase.UpdateServerStatus(server)
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		metrics.RequestsTotal.WithLabelValues("error").Inc()
	}

	proxy.ServeHTTP(w, r)
	metrics.RequestsTotal.WithLabelValues("success").Inc()
}
