CREATE TABLE IF NOT EXISTS users (
    user_id TEXT PRIMARY KEY,
    user_name TEXT,
    user_pass VARCHAR(100) NOT NULL,
    wallet NUMERIC(12,2) DEFAULT 0
);

CREATE TABLE IF NOT EXISTS stocks (
    stock_id SERIAL UNIQUE PRIMARY KEY,
    stock_name TEXT,
    current_price NUMERIC(12,2) DEFAULT 0
);

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION pass_encrypt() RETURNS TRIGGER AS $$
    BEGIN
    NEW.user_pass = crypt(NEW.user_pass, gen_salt('md5'));
    RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER login
BEFORE INSERT OR UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION pass_encrypt();