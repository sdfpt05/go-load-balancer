// test_server.go
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.Int("port", 8081, "port to listen on")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from server on port %d!\n", *port)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Test server listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
