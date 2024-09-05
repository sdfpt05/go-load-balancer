package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	defaultTimeout             = 5 * time.Second
	defaultHealthCheckInterval = 10 * time.Second
)

type Server struct {
	Address     string
	Active      bool
	Connections int64
}

type Config struct {
	ListenAddr string   `json:"listen_addr"`
	Servers    []string `json:"servers"`
	TLSCert    string   `json:"tls_cert"`
	TLSKey     string   `json:"tls_key"`
}

type LoadBalancer struct {
	servers      []*Server
	serversMutex sync.RWMutex
	config       Config
	logger       *log.Logger
}

func NewLoadBalancer(configFile string) (*LoadBalancer, error) {
	config, err := loadConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	lb := &LoadBalancer{
		config: config,
		logger: log.New(os.Stdout, "LoadBalancer: ", log.LstdFlags),
	}

	for _, addr := range config.Servers {
		lb.servers = append(lb.servers, &Server{Address: addr, Active: true})
	}

	return lb, nil
}

func loadConfig(configFile string) (Config, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func (lb *LoadBalancer) Run(ctx context.Context) error {
	ln, err := lb.createListener()
	if err != nil {
		return err
	}
	defer ln.Close()

	lb.logger.Printf("Listening for connections on %s...\n", lb.config.ListenAddr)

	go lb.checkHealthyServers(ctx)

	return lb.acceptRequests(ctx, ln)
}

func (lb *LoadBalancer) createListener() (net.Listener, error) {
	if lb.config.TLSCert != "" && lb.config.TLSKey != "" {
		cert, err := tls.LoadX509KeyPair(lb.config.TLSCert, lb.config.TLSKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
		}
		tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
		return tls.Listen("tcp", lb.config.ListenAddr, tlsConfig)
	}
	return net.Listen("tcp", lb.config.ListenAddr)
}

func (lb *LoadBalancer) acceptRequests(ctx context.Context, ln net.Listener) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			conn, err := ln.Accept()
			if err != nil {
				lb.logger.Println("Error accepting connection:", err)
				continue
			}
			go lb.handleConnection(conn)
		}
	}
}

func (lb *LoadBalancer) getNextServer() (*Server, error) {
	lb.serversMutex.RLock()
	defer lb.serversMutex.RUnlock()

	if len(lb.servers) == 0 {
		return nil, errors.New("no servers available")
	}

	var leastLoadedServer *Server
	for _, srv := range lb.servers {
		if srv.Active && (leastLoadedServer == nil || srv.Connections < leastLoadedServer.Connections) {
			leastLoadedServer = srv
		}
	}

	if leastLoadedServer == nil {
		return nil, errors.New("all servers are down")
	}

	return leastLoadedServer, nil
}

func (lb *LoadBalancer) handleConnection(conn net.Conn) {
	defer conn.Close()
	lb.logger.Printf("Received request from %s\n", conn.RemoteAddr())

	clientReq, err := readFromConnection(conn, defaultTimeout)
	if err != nil {
		lb.logger.Println("Error reading from client:", err)
		return
	}

	for retries := 3; retries > 0; retries-- {
		srv, err := lb.getNextServer()
		if err != nil {
			lb.logger.Println("Error getting next server:", err)
			sendErrorResponse(conn, "502 Bad Gateway")
			return
		}

		if err := lb.forwardRequest(srv, conn, clientReq); err == nil {
			return
		}

		lb.logger.Printf("Retry attempt %d failed\n", 4-retries)
	}

	sendErrorResponse(conn, "502 Bad Gateway")
}

func (lb *LoadBalancer) forwardRequest(srv *Server, clientConn net.Conn, clientReq []byte) error {
	lb.serversMutex.Lock()
	srv.Connections++
	lb.serversMutex.Unlock()
	defer func() {
		lb.serversMutex.Lock()
		srv.Connections--
		lb.serversMutex.Unlock()
	}()

	beConn, err := net.DialTimeout("tcp", srv.Address, defaultTimeout)
	if err != nil {
		lb.logger.Println("Error connecting to backend server:", err)
		lb.deactivateServer(srv)
		return err
	}
	defer beConn.Close()

	if _, err := beConn.Write(clientReq); err != nil {
		lb.logger.Println("Error writing to backend server:", err)
		lb.deactivateServer(srv)
		return err
	}

	backendRes, err := readFromConnection(beConn, defaultTimeout)
	if err != nil {
		lb.logger.Println("Error reading from backend server:", err)
		lb.deactivateServer(srv)
		return err
	}

	lb.logger.Printf("Response from server %s: %s", srv.Address, backendRes)
	_, err = clientConn.Write(backendRes)
	return err
}

func (lb *LoadBalancer) deactivateServer(srv *Server) {
	lb.serversMutex.Lock()
	srv.Active = false
	lb.serversMutex.Unlock()
}

func sendErrorResponse(conn net.Conn, status string) {
	resp := fmt.Sprintf("HTTP/1.1 %s\r\nConnection: close\r\n\r\n", status)
	conn.Write([]byte(resp))
}

func readFromConnection(conn net.Conn, timeout time.Duration) ([]byte, error) {
	conn.SetReadDeadline(time.Now().Add(timeout))
	reader := bufio.NewReader(conn)

	var buf bytes.Buffer
	var contentLength int64

	// Read headers
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		buf.WriteString(line)

		if strings.HasPrefix(strings.ToLower(line), "content-length:") {
			lengthStr := strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			contentLength, err = strconv.ParseInt(lengthStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid Content-Length: %w", err)
			}
		}

		if line == "\r\n" {
			break
		}
	}

	// Read body
	if contentLength > 0 {
		body := make([]byte, contentLength)
		_, err := io.ReadFull(reader, body)
		if err != nil {
			return nil, err
		}
		buf.Write(body)
	}

	return buf.Bytes(), nil
}

func (lb *LoadBalancer) checkHealthyServers(ctx context.Context) {
	ticker := time.NewTicker(defaultHealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			lb.serversMutex.Lock()
			for _, server := range lb.servers {
				if lb.isHealthy(server.Address) {
					server.Active = true
				} else {
					server.Active = false
				}
			}
			lb.serversMutex.Unlock()
		}
	}
}

func (lb *LoadBalancer) isHealthy(serverAddress string) bool {
	conn, err := net.DialTimeout("tcp", serverAddress, defaultTimeout)
	if err != nil {
		lb.logger.Printf("Could not connect to server %s: %v\n", serverAddress, err)
		return false
	}
	defer conn.Close()

	_, err = conn.Write([]byte("GET /health HTTP/1.1\r\nHost: " + serverAddress + "\r\n\r\n"))
	if err != nil {
		lb.logger.Printf("Error writing to server %s: %v\n", serverAddress, err)
		return false
	}

	resp, err := readFromConnection(conn, defaultTimeout)
	if err != nil {
		lb.logger.Printf("Error reading from server %s: %v\n", serverAddress, err)
		return false
	}

	lb.logger.Printf("Health check response from server %s: %s", serverAddress, resp)

	return strings.Contains(string(resp), "200 OK") || strings.Contains(string(resp), "204 No Content")
}

func main() {
	configFile := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	lb, err := NewLoadBalancer(*configFile)
	if err != nil {
		log.Fatalf("Failed to create load balancer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		lb.logger.Println("Received shutdown signal. Gracefully shutting down...")
		cancel()
	}()

	if err := lb.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("Load balancer error: %v", err)
	}
}
