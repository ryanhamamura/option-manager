package repository

import (
	"database/sql"
	"fmt"
	"option-manager/internal/service"
	"option-manager/internal/types"
	"time"

	_ "github.com/lib/pq"
)

// postgresRepo is the PostgreSQL implementaiton
type postgresRepo struct {
	db *sql.DB
}

// New creates a new PostgreSQL repository
func New(dataSourceName string) service.Repository {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		panic(fmt.Errorf("Failed to connect to database: %v", err))
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		panic(fmt.Errorf("Failed to ping database: %v", err))
	}

	return &postgresRepo{db: db}
}

func (r *postgresRepo) SaveUser(user types.User) (types.User, error) {
	query := `INSERT INTO users (id, email, first_name, last_name, password_hash) 
						VALUES ($1, $2, $3, $4, $5)
						RETURNING created_at, updated_at`
	var createdAt, updatedAt time.Time
	err := r.db.QueryRow(query, user.ID, user.Email, user.FirstName, user.LastName, user.PasswordHash).Scan(&createdAt, &updatedAt)
	if err != nil {
		fmt.Printf("SaveUser failed: %v\n", err)
		return types.User{}, fmt.Errorf("failed to insert user: %w", err)
	}
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt
	return user, nil
}

// GetUserByEmail retrieves User from the database with email
func (r *postgresRepo) GetUserByEmail(email string) (types.User, error) {
	var user types.User
	query := `SELECT id, email, first_name, last_name, password_hash, created_at, updated_at
						FROM users WHERE email = $1`
	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		fmt.Printf("GetUserByEmail: no user found for email %s\n", email)
		return types.User{}, fmt.Errorf("user not found: %w", err)
	}
	if err != nil {
		fmt.Printf("GetUserByEmail failed: %v\n", err)
		return types.User{}, fmt.Errorf("failed to get user: %w", err)
	}
	fmt.Printf("GetUserByEmail: succeeded: found user %+v\n", user)
	return user, nil
}

// GetPositions returns the Positions with portfolioID
func (r *postgresRepo) GetPositions(portfolioID string) ([]types.Position, error) {
	query := `SELECT id, portfolio_id, nickname, trade_price, net_liq, open_pl, closed_pl, margin_required, delta, gamma, theta, vega, created_at, updated_at 
						FROM positions WHERE portfolio_id = $1`
	rows, err := r.db.Query(query, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %v", err)
	}
	defer rows.Close()
	var positions []types.Position
	for rows.Next() {
		var p types.Position
		err := rows.Scan(&p.ID, &p.PortfolioID, &p.Nickname, &p.TradePrice, &p.NetLiq, &p.OpenPL, &p.ClosedPL, &p.MarginRequired, &p.Greeks.Delta, &p.Greeks.Gamma, &p.Greeks.Theta, &p.Greeks.Vega, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan position: %v", err)
		}
		positions = append(positions, p)
	}
	return positions, nil
}

// GetTrades returns the Trades within a Position with positionID
func (r *postgresRepo) GetTrades(positionID string) ([]types.Trade, error) {
	query := `SELECT id, position_id, nickname, trade_type, status, net_cost, margin_required, commissions, fees, open_pl, closed_pl,
										executed_at, created_at, updated_at
						FROM trades WHERE position_id = $1`
	rows, err := r.db.Query(query, positionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trades: %v", err)
	}
	defer rows.Close()
	var trades []types.Trade
	for rows.Next() {
		var t types.Trade
		err := rows.Scan(&t.ID, &t.PositionID, &t.Nickname, &t.TradeType, &t.Status, &t.NetCost, &t.MarginRequired, &t.Commissions, &t.Fees, &t.OpenPL, &t.ClosedPL, &t.ExecutedAt, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trade: %v", err)
		}
		legQuery := `SELECT id, trade_id, symbol, quantity, strike, expiration, option_type, price, current_price, delta, gamma, theta, vega 
								 FROM legs WHERE trade_id = $1`
		legRows, err := r.db.Query(legQuery, t.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get legs for trade %s: %v", t.ID, err)
		}
		defer legRows.Close()
		for legRows.Next() {
			var l types.Leg
			err := legRows.Scan(&l.ID, &l.TradeID, &l.Symbol, &l.Quantity, &l.Strike, &l.Expiration, &l.OptionType, &l.Price, &l.CurrentPrice, &l.Greeks.Delta, &l.Greeks.Gamma, &l.Greeks.Theta, &l.Greeks.Vega)
			if err != nil {
				return nil, fmt.Errorf("failed to scan leg: %v", err)
			}
			t.Legs = append(t.Legs, l)
		}
		trades = append(trades, t)
	}
	return trades, nil
}

// Close shuts down the database connection (optional, for cleanup)
func (r *postgresRepo) Close() error {
	return r.db.Close()
}
