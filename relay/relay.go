package relay

import (
	"fmt"
	"go-relay-server/config"
	"go-relay-server/queue"
	"net/smtp"
	"strings"
	"time"
)

var (
	q           *queue.Queue
	initialized bool
)

func InitializeQueue(cfg config.Config) error {
	if initialized {
		return nil
	}

	retryInterval, err := time.ParseDuration(cfg.Queue.RetryInterval)
	if err != nil {
		return fmt.Errorf("invalid retry interval: %w", err)
	}

	persistInterval, err := time.ParseDuration(cfg.Queue.PersistInterval)
	if err != nil {
		return fmt.Errorf("invalid persist interval: %w", err)
	}

	queueConfig := &queue.Config{
		StoragePath:     cfg.Queue.StoragePath,
		MaxRetries:      cfg.Queue.MaxRetries,
		RetryInterval:   retryInterval,
		MaxQueueSize:    cfg.Queue.MaxQueueSize,
		PersistInterval: persistInterval,
	}

	q, err = queue.NewQueue(queueConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize queue: %w", err)
	}

	initialized = true
	return nil
}

func RelayEmail(data []byte, from, to string, config config.Config) {
	relayServer := config.DefaultRelay
	for domain, server := range config.DomainRouting {
		if strings.Contains(to, domain) { // Fix: Use the imported `strings` package
			relayServer = server
			break
		}
	}

	fmt.Printf("Relaying email to %s: From=%s, To=%s\n", relayServer, from, to)
	err := smtp.SendMail(relayServer, nil, from, []string{to}, data)
	if err != nil {
		fmt.Printf("Failed to relay email to %s: %v\n", relayServer, err)
	} else {
		fmt.Printf("Email successfully relayed to %s\n", relayServer)
	}
}
