package storage

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/denpeshkov/receipt-processor/internal/receipt"
)

type Storage struct {
	mu sync.RWMutex
	m  map[string]*receipt.Receipt
}

func New() *Storage {
	return &Storage{m: make(map[string]*receipt.Receipt)}
}

// Store saves a new receipt.
func (s *Storage) Store(ctx context.Context, receipt *receipt.Receipt) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := uuid.New()
	receipt.ID = id.String()
	s.m[receipt.ID] = receipt
	return nil
}

// Get retrieves a  receipt by the provided ID.
func (s *Storage) Get(ctx context.Context, id string) (*receipt.Receipt, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.m[id], nil
}
