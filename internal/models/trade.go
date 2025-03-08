package models

import (
	"fmt"
	"time"
)

// TradeType represents the type of option trade
type TradeType string

const (
	// Call option trade
	Call TradeType = "CALL"
	// Put option trade
	Put TradeType = "PUT"
)

// TradeStatus represents the current status of a trade
type TradeStatus string

const (
	// Pending means the trade has been detected but not placed
	Pending TradeStatus = "PENDING"
	// Placed means the trade has been placed
	Placed TradeStatus = "PLACED"
	// Executed means the trade has been executed
	Executed TradeStatus = "EXECUTED"
	// Failed means the trade placement failed
	Failed TradeStatus = "FAILED"
)

// Trade represents an options trade from the Slack channel
type Trade struct {
	ID          string      `json:"id"`
	Symbol      string      `json:"symbol"`
	TradeType   TradeType   `json:"type"`
	Strike      float64     `json:"strike"`
	Expiration  time.Time   `json:"expiration"`
	Price       float64     `json:"price"`
	Status      TradeStatus `json:"status"`
	Description string      `json:"description"`
	RawText     string      `json:"raw_text"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// FormatThinkorswim returns the trade in Thinkorswim compatible format
func (t *Trade) FormatThinkorswim() string {
	// The raw text is already in the proper format for Thinkorswim
	return t.RawText
}

// ToMap converts a trade to a map for template rendering
func (t *Trade) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":          t.ID,
		"symbol":      t.Symbol,
		"type":        t.TradeType,
		"strike":      t.Strike,
		"expiration":  t.Expiration.Format("2006-01-02"),
		"price":       t.Price,
		"status":      t.Status,
		"description": t.Description,
		"raw_text":    t.RawText,
		"created_at":  t.CreatedAt.Format(time.RFC3339),
		"updated_at":  t.UpdatedAt.Format(time.RFC3339),
	}
}

// Validate ensures the trade has all required fields
func (t *Trade) Validate() error {
	if t.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}
	if t.TradeType != Call && t.TradeType != Put {
		return fmt.Errorf("trade type must be CALL or PUT")
	}
	if t.Strike <= 0 {
		return fmt.Errorf("strike price must be greater than 0")
	}
	if t.Expiration.IsZero() {
		return fmt.Errorf("expiration date is required")
	}
	return nil
}
