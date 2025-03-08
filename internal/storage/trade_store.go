package storage

import (
	"fmt"
	"sync"
	"time"

	"option-manager/internal/models"
)

// TradeStore defines the interface for trade storage
type TradeStore interface {
	// Get retrieves a trade by ID
	Get(id string) (*models.Trade, error)
	// List returns all trades, optionally filtered by status
	List(status *models.TradeStatus) ([]*models.Trade, error)
	// Create adds a new trade to the store
	Create(trade *models.Trade) error
	// Update updates an existing trade
	Update(trade *models.Trade) error
	// Delete removes a trade from the store
	Delete(id string) error
}

// InMemoryTradeStore is a simple in-memory implementation of TradeStore
type InMemoryTradeStore struct {
	trades map[string]*models.Trade
	mu     sync.RWMutex
}

// NewInMemoryTradeStore creates a new in-memory trade store
func NewInMemoryTradeStore() *InMemoryTradeStore {
	return &InMemoryTradeStore{
		trades: make(map[string]*models.Trade),
	}
}

// Get retrieves a trade by ID
func (s *InMemoryTradeStore) Get(id string) (*models.Trade, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	trade, ok := s.trades[id]
	if !ok {
		return nil, fmt.Errorf("trade with ID %s not found", id)
	}
	return trade, nil
}

// List returns all trades, optionally filtered by status
func (s *InMemoryTradeStore) List(status *models.TradeStatus) ([]*models.Trade, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*models.Trade
	for _, trade := range s.trades {
		if status == nil || trade.Status == *status {
			result = append(result, trade)
		}
	}
	return result, nil
}

// Create adds a new trade to the store
func (s *InMemoryTradeStore) Create(trade *models.Trade) error {
	if err := trade.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.trades[trade.ID]; exists {
		return fmt.Errorf("trade with ID %s already exists", trade.ID)
	}

	now := time.Now()
	trade.CreatedAt = now
	trade.UpdatedAt = now

	s.trades[trade.ID] = trade
	return nil
}

// Update updates an existing trade
func (s *InMemoryTradeStore) Update(trade *models.Trade) error {
	if err := trade.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existingTrade, exists := s.trades[trade.ID]
	if !exists {
		return fmt.Errorf("trade with ID %s not found", trade.ID)
	}

	// Preserve creation timestamp
	trade.CreatedAt = existingTrade.CreatedAt
	trade.UpdatedAt = time.Now()

	s.trades[trade.ID] = trade
	return nil
}

// Delete removes a trade from the store
func (s *InMemoryTradeStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.trades[id]; !exists {
		return fmt.Errorf("trade with ID %s not found", id)
	}

	delete(s.trades, id)
	return nil
}
