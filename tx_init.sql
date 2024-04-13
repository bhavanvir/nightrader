CREATE TABLE IF NOT EXISTS stock_transactions (
    stock_tx_id TEXT UNIQUE PRIMARY KEY,
    user_name TEXT,
    stock_id TEXT,
    wallet_tx_id TEXT,
    order_status TEXT,
    parent_stock_tx_id TEXT,
    is_buy BOOLEAN,
    order_type TEXT,
    stock_price NUMERIC(20,2) NOT NULL,
    quantity NUMERIC(20,2),
    time_stamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) PARTITION BY HASH(stock_tx_id);

CREATE TABLE IF NOT EXISTS stock_transactions_h0 PARTITION OF stock_transactions FOR VALUES WITH (modulus 4, remainder 0);
CREATE TABLE IF NOT EXISTS stock_transactions_h1 PARTITION OF stock_transactions FOR VALUES WITH (modulus 4, remainder 1);
CREATE TABLE IF NOT EXISTS stock_transactions_h2 PARTITION OF stock_transactions FOR VALUES WITH (modulus 4, remainder 2);
CREATE TABLE IF NOT EXISTS stock_transactions_h3 PARTITION OF stock_transactions FOR VALUES WITH (modulus 4, remainder 3);

CREATE INDEX IF NOT EXISTS stock_tx_idx ON stock_transactions USING HASH (user_name);

CREATE TABLE IF NOT EXISTS wallet_transactions (
    wallet_tx_id TEXT UNIQUE PRIMARY KEY,
    user_name TEXT,
    is_debit BOOLEAN,
    amount NUMERIC(20,2),
    time_stamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) PARTITION BY HASH(wallet_tx_id);

CREATE TABLE IF NOT EXISTS wallet_transactions_h0 PARTITION OF wallet_transactions FOR VALUES WITH (modulus 4, remainder 0);
CREATE TABLE IF NOT EXISTS wallet_transactions_h1 PARTITION OF wallet_transactions FOR VALUES WITH (modulus 4, remainder 1);
CREATE TABLE IF NOT EXISTS wallet_transactions_h2 PARTITION OF wallet_transactions FOR VALUES WITH (modulus 4, remainder 2);
CREATE TABLE IF NOT EXISTS wallet_transactions_h3 PARTITION OF wallet_transactions FOR VALUES WITH (modulus 4, remainder 3);

CREATE INDEX IF NOT EXISTS wallet_tx_idx ON wallet_transactions USING HASH (user_name);

ALTER SYSTEM SET port = 5430;