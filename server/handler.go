package server

import (
	"crypto/tls"
	"go-relay-server/config"
	"go-relay-server/logger"
	"go-relay-server/relay"
	"net"
	"net/textproto"
	"strings"
	"sync"
	"time"
)

type RateLimitingConfig struct {
	RequestsPerMinute int
	BurstLimit        int
	ExemptIPs         []string
}

type rateLimiter struct {
	requests map[string]int
	lastTime map[string]time.Time
	mu       sync.Mutex
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{
		requests: make(map[string]int),
		lastTime: make(map[string]time.Time),
	}
}

func (rl *rateLimiter) allow(ip string, config RateLimitingConfig) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Check if IP is exempt
	for _, exemptIP := range config.ExemptIPs {
		if ip == exemptIP {
			return true
		}
	}

	now := time.Now()

	// Reset counter if window has passed
	if now.Sub(rl.lastTime[ip]) > time.Minute {
		rl.requests[ip] = 0
		rl.lastTime[ip] = now
	}

	// Check rate limit
	if rl.requests[ip] >= config.RequestsPerMinute {
		return false
	}

	rl.requests[ip]++
	return true
}

func (s *Server) handleConnection(conn net.Conn, cfg config.ListenerConfig) {
	defer conn.Close()

	// Parse remote address handling both IPv4 and IPv6
	remoteAddr := conn.RemoteAddr().String()
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		s.Logger.Log(logger.LogLevelError, "Error parsing remote address %s: %v", remoteAddr, err)
		conn.Write([]byte("421 Service not available\r\n"))
		return
	}

	s.Logger.Log(logger.LogLevelInfo, "New connection from %s", host)

	// Check IP blocking
	if s.isBlocked(host) {
		s.Logger.Log(logger.LogLevelWarn, "Blocked connection from %s", host)
		conn.Write([]byte("550 Connection blocked\r\n"))
		return
	}

	// Handle STARTTLS command if configured
	if cfg.Encryption == "starttls" {
		tp := textproto.NewConn(conn)
		tp.PrintfLine("220 Welcome to the SMTP Relay Server")

		// Wait for STARTTLS command
		for {
			line, err := tp.ReadLine()
			if err != nil {
				s.Logger.Log(logger.LogLevelError, "Error reading from %s: %v", remoteAddr, err)
				return
			}

			if strings.ToUpper(line) == "STARTTLS" {
				tp.PrintfLine("220 Ready to start TLS")
				conn = tls.Server(conn, s.tlsConfig)
				s.Logger.Log(logger.LogLevelInfo, "Upgraded connection to STARTTLS from %s", remoteAddr)
				break
			}

			// Handle other commands before STARTTLS
			cmd := strings.ToUpper(strings.Fields(line)[0])
			switch cmd {
			case "HELO", "EHLO":
				tp.PrintfLine("250 Hello")
			case "QUIT":
				tp.PrintfLine("221 Bye")
				return
			default:
				tp.PrintfLine("500 Must issue STARTTLS first")
			}
		}
	} else if cfg.Encryption == "tls" {
		// Upgrade to TLS immediately for SMTPS
		conn = tls.Server(conn, s.tlsConfig)
		s.Logger.Log(logger.LogLevelInfo, "Upgraded connection to TLS from %s", remoteAddr)
	}

	// SMTP protocol handling
	tp := textproto.NewConn(conn)
	tp.PrintfLine("220 Welcome to the SMTP Relay Server")

	var from, to string
	for {
		line, err := tp.ReadLine()
		if err != nil {
			s.Logger.Log(logger.LogLevelError, "Error reading from %s: %v", remoteAddr, err)
			return
		}

		cmd := strings.ToUpper(strings.Fields(line)[0])
		switch cmd {
		case "HELO", "EHLO":
			s.Logger.Log(logger.LogLevelInfo, "Received %s command from %s", cmd, remoteAddr)
			tp.PrintfLine("250 Hello")
		case "MAIL":
			from = strings.TrimPrefix(line, "MAIL FROM:<")
			from = strings.TrimSuffix(from, ">")
			s.Logger.Log(logger.LogLevelInfo, "Received MAIL command from %s: From=%s", remoteAddr, from)
			tp.PrintfLine("250 OK")
		case "RCPT":
			to = strings.TrimPrefix(line, "RCPT TO:<")
			to = strings.TrimSuffix(to, ">")
			s.Logger.Log(logger.LogLevelInfo, "Received RCPT command from %s: To=%s", remoteAddr, to)
			if s.isBlocked(to) {
				tp.PrintfLine("550 Recipient blocked")
				s.Logger.Log(logger.LogLevelWarn, "Blocked email to %s", to)
				continue
			}
			tp.PrintfLine("250 OK")
		case "DATA":
			s.Logger.Log(logger.LogLevelInfo, "Received DATA command from %s", remoteAddr)
			tp.PrintfLine("354 Start mail input; end with <CRLF>.<CRLF>")
			data, err := tp.ReadDotBytes()
			if err != nil {
				s.Logger.Log(logger.LogLevelError, "Error reading data from %s: %v", remoteAddr, err)
				return
			}

			// Extract subject from email data
			subject := extractSubject(data)
			s.Logger.Log(logger.LogLevelInfo, "Received email from %s: From=%s, To=%s, Subject=%s", remoteAddr, from, to, subject)
			s.Logger.Log(logger.LogLevelInfo, "Email data: %s", string(data))
			relay.RelayEmail(data, from, to, s.Config)
			tp.PrintfLine("250 OK")
		case "QUIT":
			s.Logger.Log(logger.LogLevelInfo, "Received QUIT command from %s", remoteAddr)
			tp.PrintfLine("221 Bye")
			return
		default:
			s.Logger.Log(logger.LogLevelWarn, "Received unrecognized command from %s: %s", remoteAddr, line)
			tp.PrintfLine("500 Unrecognized command")
		}
	}
}

func (s *Server) isBlocked(target string) bool {
	// Parse target IP
	targetIP := net.ParseIP(target)
	if targetIP == nil {
		// Not an IP address, check as string
		for _, blocked := range s.Config.BlockList {
			if strings.Contains(target, blocked) {
				return true
			}
		}
		return false
	}

	// Check against block list
	for _, blocked := range s.Config.BlockList {
		// Try parsing as IP
		blockedIP := net.ParseIP(blocked)
		if blockedIP != nil {
			if blockedIP.Equal(targetIP) {
				return true
			}
			continue
		}

		// Try parsing as CIDR
		_, blockedNet, err := net.ParseCIDR(blocked)
		if err == nil {
			if blockedNet.Contains(targetIP) {
				return true
			}
			continue
		}

		// Fallback to string matching
		if strings.Contains(target, blocked) {
			return true
		}
	}
	return false
}

func extractSubject(data []byte) string {
	lines := strings.Split(string(data), "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.ToUpper(line), "SUBJECT:") {
			return strings.TrimSpace(line[len("Subject:"):])
		}
	}
	return "(No Subject)"
}
