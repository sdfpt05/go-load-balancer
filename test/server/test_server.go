package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

var connectionCount int64

func main() {
	port := flag.Int("port", 8081, "port to listen on")
	processingTime := flag.Duration("time", 100*time.Millisecond, "simulated processing time (e.g., 100ms, 2s)")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&connectionCount, 1)
		log.Printf("Incoming request: %s (Total connections: %d)", r.URL.String(), atomic.LoadInt64(&connectionCount))
		time.Sleep(*processingTime) // Simulate processing time
		fmt.Fprintf(w, "Hello from server on port %d! (Processing time: %v)\n", *port, *processingTime)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&connectionCount, 1)
		log.Printf("Health check request: %s (Total connections: %d)", r.URL.String(), atomic.LoadInt64(&connectionCount))
		w.WriteHeader(http.StatusOK)
	})

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Test server listening on %s (Simulated processing time: %v)\n", addr, *processingTime)
	log.Fatal(http.ListenAndServe(addr, nil))
}
