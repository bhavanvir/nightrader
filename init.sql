CREATE TABLE IF NOT EXISTS users (
    user_name TEXT PRIMARY KEY,
    name TEXT,
    user_pass VARCHAR(100) NOT NULL,
    wallet NUMERIC(12,2) DEFAULT 0
);

CREATE TABLE IF NOT EXISTS stocks (
    stock_id TEXT UNIQUE PRIMARY KEY,
    stock_name TEXT UNIQUE,
    current_price NUMERIC(12,2) DEFAULT 0,
    time_added TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_stocks (
    user_name TEXT REFERENCES users(user_name),
    stock_id TEXT REFERENCES stocks(stock_id),
    quantity INTEGER,
    PRIMARY KEY (user_name, stock_id)
);

CREATE TABLE IF NOT EXISTS wallet_transactions (
    wallet_tx_id TEXT UNIQUE PRIMARY KEY,
    user_name TEXT REFERENCES users(user_name),
    is_debit BOOLEAN,
    amount NUMERIC(12,2),
    time_stamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS stock_transactions (
    stock_tx_id TEXT UNIQUE PRIMARY KEY,
    user_name TEXT REFERENCES users(user_name),
    stock_id TEXT REFERENCES stocks(stock_id),
    wallet_tx_id TEXT,
    order_status TEXT,
    parent_stock_tx_id TEXT,
    is_buy BOOLEAN,
    order_type TEXT,
    stock_price NUMERIC(12,2) NOT NULL,
    quantity INTEGER,
    time_stamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

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
