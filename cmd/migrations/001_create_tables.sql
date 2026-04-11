-- +goose Up
CREATE TABLE IF NOT EXISTS camps(
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    address TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS shifts(
    id UUID PRIMARY KEY,
    camp_id UUID REFERENCES camps(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL
);
CREATE TABLE IF NOT EXISTS users(
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE
);
CREATE TABLE IF NOT EXISTS camp_users(
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    camp_id UUID REFERENCES camps(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    PRIMARY KEY (user_id, camp_id)
);

-- +goose Down
DROP TABLE IF EXISTS shifts;
DROP TABLE IF EXISTS camp_users;
DROP TABLE IF EXISTS camps;
DROP TABLE IF EXISTS users;