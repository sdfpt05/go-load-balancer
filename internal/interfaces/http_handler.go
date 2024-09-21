package interfaces

import (
	"context"
	"github.com/sdfpt05/go_load_balancer/v2/internal/usecases"
	"log"
	"net/http"
	"net/http/httputil"
	"sync/atomic"
	"time"
)

type HTTPHandler struct {
	loadBalancerUseCase *usecases.LoadBalancerUseCase
}

func NewHTTPHandler(uc *usecases.LoadBalancerUseCase) *HTTPHandler {
	return &HTTPHandler{loadBalancerUseCase: uc}
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	log.Printf("Incoming request:\n%s", string(r.URL.String()))

	server, err := h.loadBalancerUseCase.GetNextServer(ctx)
	if err != nil {
		http.Error(w, "No server available", http.StatusServiceUnavailable)
		log.Printf("Request failed: No server available")
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(server.URL)

	// Create a custom response writer to capture the status code and response
	rw := &responseWriter{ResponseWriter: w}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, "Server error", http.StatusBadGateway)
		server.Active.Store(false)
		h.loadBalancerUseCase.UpdateServerStatus(server)
		log.Printf("Request failed: Server error - %v", err)
	}

	start := time.Now()
	proxy.ServeHTTP(rw, r)
	duration := time.Since(start)

	server.ResponseTime = duration
	atomic.AddInt64(&server.Connections, 1)
	h.loadBalancerUseCase.UpdateServerStatus(server)

	if rw.statusCode >= 200 && rw.statusCode < 300 {
		log.Printf("Request successful:\n%s", string(rw.statusCode))
	} else {
		log.Printf("Request failed:\n%s", string(rw.statusCode))
	}
}

// Custom ResponseWriter to capture the status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}
