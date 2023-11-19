CREATE TABLE IF NOT EXISTS users
(
    id           SERIAL PRIMARY KEY,
    first_name   VARCHAR(255),
    last_name    VARCHAR(255),
    user_active  INTEGER DEFAULT 0,
    access_level INTEGER DEFAULT 3,
    email        VARCHAR(255),
    password     VARCHAR(60),
    deleted_at   TIMESTAMP,
    created_at   TIMESTAMP,
    updated_at   TIMESTAMP
);

ALTER TABLE users
    ALTER COLUMN created_at SET DEFAULT now();
ALTER TABLE users
    ALTER COLUMN updated_at SET DEFAULT now();

CREATE TRIGGER set_timestamp
    BEFORE UPDATE
    ON users
    FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();