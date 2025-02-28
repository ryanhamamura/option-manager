CREATE TABLE portfolios (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_portfolios_timestamp
    BEFORE UPDATE ON portfolios
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TABLE positions (
    id UUID PRIMARY KEY,
    portfolio_id UUID REFERENCES portfolios(id),
    nickname VARCHAR(50) NOT NULL,
    trade_price DOUBLE PRECISION DEFAULT 0.0,
    net_liq DOUBLE PRECISION DEFAULT 0.0,
    open_pl DOUBLE PRECISION DEFAULT 0.0,
    closed_pl DOUBLE PRECISION DEFAULT 0.0,
    margin_required DOUBLE PRECISION DEFAULT 0.0,
    delta DOUBLE PRECISION DEFAULT 0.0,
    gamma DOUBLE PRECISION DEFAULT 0.0,
    theta DOUBLE PRECISION DEFAULT 0.0,
    vega DOUBLE PRECISION DEFAULT 0.0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_positions_timestamp
    BEFORE UPDATE ON positions
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TABLE trades (
    id UUID PRIMARY KEY,
    position_id UUID REFERENCES positions(id),
    nickname VARCHAR(50) NOT NULL,
    trade_type VARCHAR(50) NOT NULL,
    status VARCHAR(10) NOT NULL CHECK (status IN ('open', 'closed', 'proposed')) DEFAULT 'open',
    net_cost DOUBLE PRECISION NOT NULL,
    margin_required DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    commissions DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    fees DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    open_pl DOUBLE PRECISION DEFAULT 0.0,
    closed_pl DOUBLE PRECISION DEFAULT 0.0,
    executed_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_trades_timestamp
    BEFORE UPDATE ON trades
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TABLE legs (
    id UUID PRIMARY KEY,
    trade_id UUID REFERENCES trades(id),
    symbol VARCHAR(50) NOT NULL,
    quantity INT NOT NULL,
    strike DOUBLE PRECISION NOT NULL,
    expiration TIMESTAMP NOT NULL,
    option_type VARCHAR(4) CHECK (option_type IN ('call', 'put')),
    price DOUBLE PRECISION NOT NULL,
    current_price DOUBLE PRECISION DEFAULT 0.0,
    delta DOUBLE PRECISION DEFAULT 0.0,
    gamma DOUBLE PRECISION DEFAULT 0.0,
    theta DOUBLE PRECISION DEFAULT 0.0,
    vega DOUBLE PRECISION DEFAULT 0.0
);

CREATE INDEX idx_positions_portfolio_id ON positions(portfolio_id);
CREATE INDEX idx_trades_position_id ON trades(position_id);
CREATE INDEX idx_legs_trade_id ON legs(trade_id);
