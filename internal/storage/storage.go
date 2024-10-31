// Package storage provides an in-memory storage for receipts.
package storage

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/denpeshkov/receipt-processor/internal/receipt"
)

// Storage is an in-memory storage for receipts. It is safe for use by multiple goroutines.
type Storage struct {
	mu sync.RWMutex
	m  map[string]*receipt.Receipt
}

// New returns a new storage.
func New() *Storage {
	return &Storage{m: make(map[string]*receipt.Receipt)}
}

// Store saves a new receipt in the storage, generating a unique ID for it.
func (s *Storage) Store(ctx context.Context, receipt *receipt.Receipt) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := uuid.New()
	receipt.ID = id.String()
	s.m[receipt.ID] = receipt
	return nil
}

// Get retrieves a receipt from the storage by its ID.
func (s *Storage) Get(ctx context.Context, id string) (*receipt.Receipt, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.m[id], nil
}
