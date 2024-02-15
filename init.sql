CREATE TABLE IF NOT EXISTS users (
    user_name TEXT PRIMARY KEY,
    name TEXT,
    user_pass VARCHAR(100) NOT NULL,
    wallet NUMERIC(12,2) DEFAULT 0
);

CREATE TABLE IF NOT EXISTS stocks (
    stock_id SERIAL UNIQUE PRIMARY KEY,
    stock_name TEXT,
    current_price NUMERIC(12,2) DEFAULT 0
);

CREATE TABLE IF NOT EXISTS user_stocks (
    user_name TEXT REFERENCES users(user_name),
    stock_id INTEGER REFERENCES stocks(stock_id),
    quantity INTEGER,
    PRIMARY KEY (user_name, stock_id)
);

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION pass_encrypt() RETURNS TRIGGER AS $$
    BEGIN
    NEW.user_pass = crypt(NEW.user_pass, gen_salt('md5'));
    RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER register
BEFORE INSERT OR UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION pass_encrypt();

CREATE OR REPLACE FUNCTION login(u_name text, u_pass text) RETURNS boolean AS $$
    DECLARE
    is_valid boolean;
    BEGIN
    SELECT user_pass = crypt(u_pass, user_pass) INTO STRICT is_valid
        FROM users WHERE user_name = u_name;
    RETURN is_valid;
    EXCEPTION
        WHEN NO_DATA_FOUND THEN
            RAISE EXCEPTION 'user % not found', u_name;
        WHEN TOO_MANY_ROWS THEN
            RAISE EXCEPTION 'user % not unique', u_name;
    END
$$ LANGUAGE plpgsql;
