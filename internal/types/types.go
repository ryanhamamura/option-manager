package types

import "time"

// Portfolio represents a user's collection of positions
type Portfolio struct {
	ID        string
	UserID    string // Links to User.ID
	Name      string // e.g., "Tech Portfolio" or "March 2025 Plays"
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Position represents a user-defined collection of trades within a portfolio, aggregating high-level metrics across multiple strategies and underlyings
type Position struct {
	ID             string
	PortfolioID    string  // Links to Portfolio.ID
	Nickname       string  // User-defined, e.g., "SPX-XSP Hedge Strategy", "Rhino"
	TradePrice     float64 // Weighted average price across all trades (computed)
	NetLiq         float64 // Net liquidating value of all trades (current value)
	OpenPL         float64 // Aggregated unrealized P/L from all open trades
	ClosedPL       float64 // Aggregated realized P/L from all closed trades
	MarginRequired float64 // Total margin required across all trades (computed)
	Greeks         Greeks  // Aggregate Greeks across all trades
	Trades         []Trade
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Trade represents a single transaction within a position
// Can involve up to 4 option legs (e.g., spread, iron condor)
type Trade struct {
	ID             string
	PositionID     string    // Links to Position.ID
	Nickname       string    // User-defined or default, e.g., "Butterfly #24"
	TradeType      string    // e.g., "iron condor", "butterfly", "calendar"
	Status         string    // "open", "closed", "proposed"
	Legs           []Leg     // Up to 4 option legs
	NetCost        float64   // Net debit (positive) or credit (negative) for the trade
	MarginRequired float64   // Margin required for defined-risk trades
	Commissions    float64   // Commission costs for the trade (e.g., $1 per contract)
	Fees           float64   // Additional fees (e.g., SEC, clearing fees)
	OpenPL         float64   // Unrealized P/L while open
	ClosedPL       float64   // Realized P/L once closed
	ExecutedAt     time.Time // When the trade was placed (relevant for open/closed)
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Leg represents a single option leg within a trade
type Leg struct {
	ID           string
	TradeID      string // Links to Trade.ID
	Symbol       string // e.g., "SPX 250321C150" or "XSP 250321C150"
	Quantity     int    // Positive for long, negative for short
	Strike       float64
	Expiration   time.Time
	OptionType   string  // "call" or "put"
	Price        float64 // Price at execution (per contract)
	CurrentPrice float64 // Latest quote for this leg
	Greeks       Greeks  // Greeks specific to this leg
}

// Greeks holds options Greeks for a leg or aggregated position
type Greeks struct {
	Delta float64
	Gamma float64
	Theta float64
	Vega  float64
}

// Quote represents real-time market data for an option or underlying (in-memory)
type Quote struct {
	Symbol     string    // e.g., "SPX 250321C150" or "XSP"
	Strike     float64   // Strike price (for options)
	Expiration time.Time // Expiration date (for options)
	OptionType string    // "call" or "put" (for options)
	Bid        float64   // Bid price
	Ask        float64   // Ask price
	Last       float64   // Last traded price
	Mid        float64   // Midpoint of bid/ask
	Exchange   string    // e.g., "CBOE", "NASDAQ"
	Greeks     Greeks    // Greeks for this quote (if applicable)
	UpdatedAt  time.Time // Last update timestamp
}

// User represents a user in the system
type User struct {
	ID                 string
	Email              string
	PasswordHash       string
	FirstName          string
	LastName           string
	EmailVerified      bool
	VerificationToken  *string
	VerificationExpiry *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// Session represents the session model
type Session struct {
	ID        string
	UserID    int
	ExpiresAt time.Time
	CreatedAt time.Time
}
