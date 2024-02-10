CREATE TABLE IF NOT EXISTS users (
    user_name text primary key,
    name text,
    password text,
    wallet integer
);

CREATE TABLE IF NOT EXISTS stocks (
    stock_id serial unique primary key,
    stock_name text,
    current_price integer
);