package interfaces

import (
	"context"
	"myapp/usecases"
	"net/http"
	"net/http/httputil"
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

	server, err := h.loadBalancerUseCase.GetNextServer(ctx)
	if err != nil {
		http.Error(w, "No server available", http.StatusServiceUnavailable)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(server.URL)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, "Server error", http.StatusBadGateway)
		server.Active.Store(false)
		h.loadBalancerUseCase.UpdateServerStatus(server)
	}

	start := time.Now()
	proxy.ServeHTTP(w, r)
	duration := time.Since(start)

	server.ResponseTime = duration
	atomic.AddInt64(&server.Connections, 1)
	h.loadBalancerUseCase.UpdateServerStatus(server)
}
