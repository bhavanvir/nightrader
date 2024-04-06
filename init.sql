CREATE TABLE IF NOT EXISTS users (
    user_name TEXT PRIMARY KEY,
    name TEXT,
    user_pass VARCHAR(100) NOT NULL,
    wallet NUMERIC(20,2) DEFAULT 0
);

CREATE TABLE IF NOT EXISTS stocks (
    stock_id TEXT UNIQUE PRIMARY KEY,
    stock_name TEXT UNIQUE,
    current_price NUMERIC(20,2) DEFAULT 0,
    time_added TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_stocks (
    user_name TEXT REFERENCES users(user_name),
    stock_id TEXT REFERENCES stocks(stock_id),
    quantity NUMERIC(20,2),
    time_added TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_name, stock_id)
);

CREATE TABLE IF NOT EXISTS wallet_transactions (
    wallet_tx_id TEXT UNIQUE PRIMARY KEY,
    user_name TEXT REFERENCES users(user_name),
    is_debit BOOLEAN,
    amount NUMERIC(20,2),
    time_stamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) PARTITION BY HASH(wallet_tx_id);

CREATE TABLE IF NOT EXISTS wallet_transactions_h0 PARTITION OF wallet_transactions FOR VALUES WITH (modulus 4, remainder 0);
CREATE TABLE IF NOT EXISTS wallet_transactions_h1 PARTITION OF wallet_transactions FOR VALUES WITH (modulus 4, remainder 1);
CREATE TABLE IF NOT EXISTS wallet_transactions_h2 PARTITION OF wallet_transactions FOR VALUES WITH (modulus 4, remainder 2);
CREATE TABLE IF NOT EXISTS wallet_transactions_h3 PARTITION OF wallet_transactions FOR VALUES WITH (modulus 4, remainder 3);

CREATE INDEX IF NOT EXISTS wallet_tx_idx ON wallet_transactions USING HASH (user_name);

CREATE TABLE IF NOT EXISTS stock_transactions (
    stock_tx_id TEXT UNIQUE PRIMARY KEY,
    user_name TEXT REFERENCES users(user_name),
    stock_id TEXT REFERENCES stocks(stock_id),
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

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION pass_encrypt() RETURNS TRIGGER AS $$
    BEGIN
    -- Check if user_pass is being updated or newly inserted
    IF TG_OP = 'INSERT' OR NEW.user_pass IS DISTINCT FROM OLD.user_pass THEN
        NEW.user_pass = crypt(NEW.user_pass, gen_salt('md5'));
    END IF;
    RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER login
BEFORE INSERT OR UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION pass_encrypt();
