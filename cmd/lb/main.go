package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultTimeout             = 5 * time.Second
	defaultHealthCheckInterval = 10 * time.Second
)

type server struct {
	address     string
	active      bool
	connections int
}

type Config struct {
	ListenAddr string   `json:"listen_addr"`
	Servers    []string `json:"servers"`
	TLSCert    string   `json:"tls_cert"`
	TLSKey     string   `json:"tls_key"`
}

var (
	servers      []*server
	serversMutex sync.RWMutex
	configFile   string
	config       Config
)

func init() {
	flag.StringVar(&configFile, "config", "config.json", "Path to configuration file")
	flag.Parse()

	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	for _, addr := range config.Servers {
		servers = append(servers, &server{address: addr, active: true})
	}
}

func loadConfig() error {
	file, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&config)
}

func main() {
	var ln net.Listener
	var err error

	if config.TLSCert != "" && config.TLSKey != "" {
		cert, err := tls.LoadX509KeyPair(config.TLSCert, config.TLSKey)
		if err != nil {
			log.Fatalf("Failed to load TLS certificate: %v", err)
		}
		tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
		ln, err = tls.Listen("tcp", config.ListenAddr, tlsConfig)
	} else {
		ln, err = net.Listen("tcp", config.ListenAddr)
	}

	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stdout, "Listening for connections on %s...\n", config.ListenAddr)

	go checkHealthyServers()
	acceptRequests(ln)
}

func acceptRequests(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func getNextServer() (*server, error) {
	serversMutex.RLock()
	defer serversMutex.RUnlock()

	if len(servers) == 0 {
		return nil, errors.New("no servers available")
	}

	leastConnections := servers[0].connections
	leastLoadedServer := servers[0]

	for _, srv := range servers {
		if srv.active && srv.connections < leastConnections {
			leastConnections = srv.connections
			leastLoadedServer = srv
		}
	}

	if !leastLoadedServer.active {
		return nil, errors.New("all servers are down")
	}

	return leastLoadedServer, nil
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Fprintf(os.Stdout, "Received request from %s\n", conn.RemoteAddr())

	clientRes, err := readFromConnection(conn)
	if err != nil {
		log.Println("Error reading from client:", err)
		return
	}

	retries := 3
	for retries > 0 {
		srv, err := getNextServer()
		if err != nil {
			log.Println("Error getting next server:", err)
			sendErrorResponse(conn, "502 Bad Gateway")
			return
		}

		serversMutex.Lock()
		srv.connections++
		serversMutex.Unlock()

		beConn, err := net.DialTimeout("tcp", srv.address, defaultTimeout)
		if err != nil {
			log.Println("Error connecting to backend server:", err)
			serversMutex.Lock()
			srv.deactivate()
			srv.connections--
			serversMutex.Unlock()
			retries--
			continue
		}

		_, err = beConn.Write([]byte(clientRes))
		if err != nil {
			log.Println("Error writing to backend server:", err)
			serversMutex.Lock()
			srv.deactivate()
			srv.connections--
			serversMutex.Unlock()
			beConn.Close()
			retries--
			continue
		}

		backendRes, err := readFromConnection(beConn)
		beConn.Close()

		serversMutex.Lock()
		srv.connections--
		serversMutex.Unlock()

		if err != nil {
			log.Println("Error reading from backend server:", err)
			serversMutex.Lock()
			srv.deactivate()
			serversMutex.Unlock()
			retries--
			continue
		}

		fmt.Fprintf(os.Stdout, "Response from server %s: %s", srv.address, backendRes)
		conn.Write([]byte(backendRes))
		return
	}

	sendErrorResponse(conn, "502 Bad Gateway")
}

func sendErrorResponse(conn net.Conn, status string) {
	resp := fmt.Sprintf("HTTP/1.1 %s\r\nConnection: close\r\n\r\n", status)
	conn.Write([]byte(resp))
}

func readFromConnection(conn net.Conn) (string, error) {
	conn.SetReadDeadline(time.Now().Add(defaultTimeout))
	reader := bufio.NewReader(conn)

	buf := bytes.Buffer{}
	contentLength := 0
	for {
		s, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		buf.WriteString(s)

		if strings.HasPrefix(s, "Content-Length:") {
			lengthInStr := strings.Split(s, ":")[1]
			contentLength, err = strconv.Atoi(strings.TrimSpace(lengthInStr))
			if err != nil {
				return "", fmt.Errorf("invalid Content-Length: %v", err)
			}
		}

		if s == "\r\n" {
			break
		}
	}

	for contentLength > 0 {
		b, err := reader.ReadByte()
		if err != nil {
			return "", err
		}
		buf.WriteByte(b)
		contentLength--
	}

	return buf.String(), nil
}

func checkHealthyServers() {
	for {
		time.Sleep(defaultHealthCheckInterval)
		serversMutex.Lock()
		for _, server := range servers {
			if isHealthy(server.address) {
				server.activate()
			} else {
				server.deactivate()
			}
		}
		serversMutex.Unlock()
	}
}

func (s *server) activate() {
	s.active = true
}

func (s *server) deactivate() {
	s.active = false
}

func isHealthy(serverAddress string) bool {
	conn, err := net.DialTimeout("tcp", serverAddress, defaultTimeout)
	if err != nil {
		fmt.Printf("Could not connect to server %s: %v\n", serverAddress, err)
		return false
	}
	defer conn.Close()

	_, err = conn.Write([]byte("GET /health HTTP/1.1\r\nHost: " + serverAddress + "\r\n\r\n"))
	if err != nil {
		fmt.Printf("Error writing to server %s: %v\n", serverAddress, err)
		return false
	}

	resp, err := readFromConnection(conn)
	if err != nil {
		fmt.Printf("Error reading from server %s: %v\n", serverAddress, err)
		return false
	}

	fmt.Printf("Health check response from server %s: %s", serverAddress, resp)

	return strings.Contains(resp, "200 OK") || strings.Contains(resp, "204 No Content")
}
