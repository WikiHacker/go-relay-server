package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
)

type ListenerConfig struct {
	Host        string `json:"host"`
	Port        string `json:"port"`
	Encryption  string `json:"encryption"`   // "none", "tls", or "starttls"
	RequireAuth bool   `json:"require_auth"` // Whether to require authentication
}

type Config struct {
	Listeners     []ListenerConfig  `json:"listeners"`
	DefaultRelay  string            `json:"default_relay"`
	AllowList     []string          `json:"allow_list"`
	BlockList     []string          `json:"block_list"`
	DomainRouting map[string]string `json:"domain_routing"`
	TLSCertFile   string            `json:"tls_cert_file"`
	TLSKeyFile    string            `json:"tls_key_file"`
	AuthUsername  string            `json:"auth_username"`
	AuthPassword  string            `json:"auth_password"`
	LogFile       string            `json:"log_file"`
	LogLevel      string            `json:"log_level"`
	RateLimiting  RateLimiting      `json:"rate_limiting"`
	Queue         QueueConfig       `json:"queue"`
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
	if len(config.Listeners) == 0 {
		return errors.New("at least one listener configuration is required")
	}

	if config.DefaultRelay == "" {
		return errors.New("default_relay is required")
	}

	// Validate listeners
	for _, listener := range config.Listeners {
		port, err := strconv.Atoi(listener.Port)
		if err != nil || port <= 0 || port > 65535 {
			return errors.New("listener port must be a valid number between 1 and 65535")
		}
		if listener.Encryption != "none" && listener.Encryption != "tls" && listener.Encryption != "starttls" {
			return errors.New("listener encryption must be one of: none, tls, starttls")
		}
		if (listener.Encryption == "tls" || listener.Encryption == "starttls") &&
			(config.TLSCertFile == "" || config.TLSKeyFile == "") {
			return errors.New("tls_cert_file and tls_key_file are required for encrypted listeners")
		}
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
