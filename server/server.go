package server

import (
	"crypto/tls"
	"fmt"
	"go-relay-server/config"
	"go-relay-server/logger"
	"net"
	"sync"
)

type Server struct {
	Config    config.Config
	Logger    *logger.Logger
	wg        sync.WaitGroup
	quit      chan struct{}
	running   bool
	mu        sync.RWMutex
	listeners []net.Listener
	tlsConfig *tls.Config
}

func NewServer(config config.Config) (*Server, error) {
	server := &Server{
		Config: config,
		quit:   make(chan struct{}),
	}

	// Convert config.LogLevel to logger.LogLevel
	logLevel := logger.LogLevel(config.LogLevel)

	// Initialize the logger
	loggerInstance, err := logger.NewLogger(config.LogFile, logLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to setup logger: %v", err)
	}
	server.Logger = loggerInstance

	// Start log rotation
	go server.Logger.DailyLogRotation()

	return server, nil
}

func (s *Server) loadTLSConfig() error {
	cert, err := tls.LoadX509KeyPair(s.Config.TLSCertFile, s.Config.TLSKeyFile)
	if err != nil {
		return fmt.Errorf("failed to load TLS certificate: %v", err)
	}

	s.tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}
	return nil
}

func (s *Server) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server is already running")
	}
	s.running = true
	s.mu.Unlock()

	// Load TLS config if needed
	for _, listenerCfg := range s.Config.Listeners {
		if listenerCfg.Encryption == "tls" || listenerCfg.Encryption == "starttls" {
			if err := s.loadTLSConfig(); err != nil {
				return err
			}
			break
		}
	}

	// Start listeners
	for _, listenerCfg := range s.Config.Listeners {
		listener, err := s.createListener(listenerCfg)
		if err != nil {
			s.Stop()
			return fmt.Errorf("failed to start listener on port %s: %v", listenerCfg.Port, err)
		}
		s.listeners = append(s.listeners, listener)
		go s.acceptConnections(listener, listenerCfg)
	}

	return nil
}

func (s *Server) createListener(cfg config.ListenerConfig) (net.Listener, error) {
	// Listen on all interfaces for both IPv4 and IPv6
	addr := ":" + cfg.Port
	network := "tcp"

	// Try listening on dual stack first
	listener, err := net.Listen(network, addr)
	if err == nil {
		return listener, nil
	}

	// If dual stack fails, try IPv4 only
	network = "tcp4"
	listener, err = net.Listen(network, addr)
	if err == nil {
		return listener, nil
	}

	// If IPv4 fails, try IPv6 only
	network = "tcp6"
	listener, err = net.Listen(network, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %v", err)
	}

	if cfg.Encryption == "tls" {
		return tls.Listen(network, addr, s.tlsConfig)
	}
	return listener, nil
}

func (s *Server) acceptConnections(listener net.Listener, cfg config.ListenerConfig) {
	s.Logger.Log(logger.LogLevelInfo, "Server started on port %s (%s)", cfg.Port, cfg.Encryption)

	for {
		select {
		case <-s.quit:
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				s.Logger.Log(logger.LogLevelError, "Error accepting connection on port %s: %v", cfg.Port, err)
				continue
			}

			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				s.handleConnection(conn, cfg)
			}()
		}
	}
}

func (s *Server) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.quit)

	// Close all listeners
	for _, listener := range s.listeners {
		listener.Close()
	}

	s.wg.Wait()
	s.listeners = nil
	s.Logger.Log(logger.LogLevelInfo, "Server stopped")
}

func (s *Server) Status() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.running {
		return "running"
	}
	return "stopped"
}

func (s *Server) Restart() error {
	s.Stop()

	// Reset quit channel and running state
	s.mu.Lock()
	s.quit = make(chan struct{})
	s.running = false
	s.mu.Unlock()

	return s.Start()
}
