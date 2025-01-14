package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

type Config struct {
	StoragePath     string
	MaxRetries      int
	RetryInterval   time.Duration
	MaxQueueSize    int
	PersistInterval time.Duration
}

type Queue struct {
	items           []*QueueItem
	failedItems     []FailedItem
	storagePath     string
	maxRetries      int
	retryInterval   time.Duration
	maxQueueSize    int
	persistTimer    *time.Timer
	persistChannel  chan struct{}
	persistInterval time.Duration
	mu              sync.Mutex
}

type FailedItem struct {
	Item      *QueueItem
	Error     string
	Timestamp time.Time
	Retries   int
}

type QueueItem struct {
	ID        string
	Data      []byte
	Attempts  int
	NextRetry time.Time
	CreatedAt time.Time
}

func NewQueue(config *Config) (*Queue, error) {
	if err := os.MkdirAll(config.StoragePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create queue storage: %w", err)
	}

	q := &Queue{
		storagePath:     config.StoragePath,
		maxRetries:      config.MaxRetries,
		retryInterval:   config.RetryInterval,
		maxQueueSize:    config.MaxQueueSize,
		persistInterval: config.PersistInterval,
		items:           make([]*QueueItem, 0),
	}

	if err := q.loadFromDisk(); err != nil {
		return nil, fmt.Errorf("failed to load queue from disk: %w", err)
	}

	go q.startPersistWorker()

	return q, nil
}

func (q *Queue) Enqueue(data []byte) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) >= q.maxQueueSize {
		return errors.New("queue is full")
	}

	item := &QueueItem{
		ID:        generateID(),
		Data:      data,
		Attempts:  0,
		NextRetry: time.Now(),
		CreatedAt: time.Now(),
	}

	q.items = append(q.items, item)
	return nil
}

func (q *Queue) startPersistWorker() {
	ticker := time.NewTicker(q.persistInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := q.persistToDisk(); err != nil {
			// Log error
		}
	}
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func (q *Queue) Dequeue() (*QueueItem, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()
	for i, item := range q.items {
		if item.NextRetry.Before(now) {
			q.items = append(q.items[:i], q.items[i+1:]...)
			return item, nil
		}
	}

	return nil, errors.New("no items ready for processing")
}

func (q *Queue) Retry(item *QueueItem) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if item.Attempts >= q.maxRetries {
		q.failedItems = append(q.failedItems, FailedItem{
			Item:      item,
			Error:     "max retries exceeded",
			Timestamp: time.Now(),
			Retries:   item.Attempts,
		})
		return errors.New("max retries exceeded")
	}

	item.Attempts++
	item.NextRetry = time.Now().Add(q.retryInterval)
	q.items = append(q.items, item)
	return nil
}

func (q *Queue) GetFailedItems() []FailedItem {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.failedItems
}

func (q *Queue) ClearFailedItems() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.failedItems = []FailedItem{}
}

func (q *Queue) RequeueFailedItem(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, failedItem := range q.failedItems {
		if failedItem.Item.ID == id {
			// Reset attempts and retry immediately
			failedItem.Item.Attempts = 0
			failedItem.Item.NextRetry = time.Now()
			q.items = append(q.items, failedItem.Item)
			q.failedItems = append(q.failedItems[:i], q.failedItems[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("failed item with ID %s not found", id)
}

func (q *Queue) loadFromDisk() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Load regular queue items
	itemsFile := fmt.Sprintf("%s/items.dat", q.storagePath)
	if _, err := os.Stat(itemsFile); err == nil {
		data, err := os.ReadFile(itemsFile)
		if err != nil {
			return fmt.Errorf("failed to read queue items: %w", err)
		}
		if err := json.Unmarshal(data, &q.items); err != nil {
			return fmt.Errorf("failed to decode queue items: %w", err)
		}
	}

	// Load failed items
	failedFile := fmt.Sprintf("%s/failed_items.dat", q.storagePath)
	if _, err := os.Stat(failedFile); err == nil {
		data, err := os.ReadFile(failedFile)
		if err != nil {
			return fmt.Errorf("failed to read failed items: %w", err)
		}
		if err := json.Unmarshal(data, &q.failedItems); err != nil {
			return fmt.Errorf("failed to decode failed items: %w", err)
		}
	}

	return nil
}

func (q *Queue) persistToDisk() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(q.storagePath, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Save regular queue items
	itemsFile := fmt.Sprintf("%s/items.dat", q.storagePath)
	itemsData, err := json.Marshal(q.items)
	if err != nil {
		return fmt.Errorf("failed to encode queue items: %w", err)
	}
	if err := os.WriteFile(itemsFile, itemsData, 0644); err != nil {
		return fmt.Errorf("failed to save queue items: %w", err)
	}

	// Save failed items
	failedFile := fmt.Sprintf("%s/failed_items.dat", q.storagePath)
	failedData, err := json.Marshal(q.failedItems)
	if err != nil {
		return fmt.Errorf("failed to encode failed items: %w", err)
	}
	if err := os.WriteFile(failedFile, failedData, 0644); err != nil {
		return fmt.Errorf("failed to save failed items: %w", err)
	}

	return nil
}
