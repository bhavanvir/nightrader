CREATE TABLE IF NOT EXISTS stocks (
    stock_id TEXT UNIQUE PRIMARY KEY,
    stock_name TEXT UNIQUE,
    current_price NUMERIC(20,2) DEFAULT 0,
    time_added TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_stocks (
    user_name TEXT,
    stock_id TEXT REFERENCES stocks(stock_id),
    quantity NUMERIC(20,2),
    time_added TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_name, stock_id)
);

ALTER SYSTEM SET port = 5431;