package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Config struct {
	ListenPort    string
	DefaultRelay  string
	AllowList     []string
	BlockList     []string
	DomainRouting map[string]string
	EnableTLS     bool
	TLSCertFile   string
	TLSKeyFile    string
	EnableAuth    bool
	AuthUsername  string
	AuthPassword  string
	LogFile       string
	LogLevel      string
	RateLimiting  RateLimiting
	Queue         QueueConfig
}

type QueueConfig struct {
	StoragePath     string
	MaxRetries      int
	RetryInterval   string
	MaxQueueSize    int
	PersistInterval string
}

type RateLimiting struct {
	RequestsPerMinute int      `json:"requests_per_minute"`
	BurstLimit        int      `json:"burst_limit"`
	ExemptIPs         []string `json:"exempt_ips"`
}

type LogLevel string

const (
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

func LoadConfig(filename string) (Config, error) {
	var config Config
	file, err := os.Open(filename)
	if err != nil {
		return config, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return config, fmt.Errorf("failed to decode config file: %v", err)
	}

	if err := validateConfig(config); err != nil {
		return config, fmt.Errorf("invalid configuration: %v", err)
	}

	return config, nil
}

func validateConfig(config Config) error {
	if config.ListenPort == "" {
		return errors.New("listen_port is required")
	}
	if config.DefaultRelay == "" {
		return errors.New("default_relay is required")
	}
	if config.EnableTLS && (config.TLSCertFile == "" || config.TLSKeyFile == "") {
		return errors.New("tls_cert_file and tls_key_file are required when enable_tls is true")
	}

	// Validate rate limiting configuration
	if config.RateLimiting.RequestsPerMinute <= 0 {
		return errors.New("rate_limiting.requests_per_minute must be positive")
	}
	if config.RateLimiting.BurstLimit <= 0 {
		return errors.New("rate_limiting.burst_limit must be positive")
	}
	if config.RateLimiting.BurstLimit > config.RateLimiting.RequestsPerMinute {
		return errors.New("rate_limiting.burst_limit cannot be greater than requests_per_minute")
	}

	return nil
}
