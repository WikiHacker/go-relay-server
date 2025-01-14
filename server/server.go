package server

import (
	"fmt"
	"go-relay-server/config"
	"go-relay-server/logger"
	"net"
	"sync"
)

type Server struct {
	Config  config.Config
	Logger  *logger.Logger
	wg      sync.WaitGroup
	quit    chan struct{}
	running bool
	mu      sync.RWMutex
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

func (s *Server) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server is already running")
	}
	s.running = true
	s.mu.Unlock()

	listener, err := net.Listen("tcp", ":"+s.Config.ListenPort)
	if err != nil {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		return fmt.Errorf("failed to start server: %v", err)
	}
	defer listener.Close()

	s.Logger.Log(logger.LogLevelInfo, "Server started on port %s", s.Config.ListenPort)

	for {
		select {
		case <-s.quit:
			s.Logger.Log(logger.LogLevelInfo, "Server shutting down...")
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				s.Logger.Log(logger.LogLevelError, "Error accepting connection: %v", err)
				continue
			}

			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				s.handleConnection(conn)
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
	s.wg.Wait()
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
