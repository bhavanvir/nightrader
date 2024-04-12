CREATE TABLE IF NOT EXISTS users (
    user_name TEXT PRIMARY KEY,
    name TEXT,
    user_pass VARCHAR(100) NOT NULL,
    wallet NUMERIC(20,2) DEFAULT 0
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