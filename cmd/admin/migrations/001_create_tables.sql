-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;
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
    description TEXT NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    price INT NOT NULL,
    capacity INT NOT NULL,
    count_bookings INT NOT NULL DEFAULT 0 ,
    status INT NOT NULL

);
CREATE TABLE IF NOT EXISTS users(
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL,
    name TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS camp_users(
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    camp_id UUID REFERENCES camps(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    PRIMARY KEY (user_id, camp_id)
);
CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    shift_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (user_id, shift_id)
);
CREATE TABLE IF NOT EXISTS participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name VARCHAR(255) NOT NULL,
    gender CHAR(1) NOT NULL,
    birth_date DATE NOT NULL,
    age INT NOT NULL,
    parent_name VARCHAR(255),
    parent_phone VARCHAR(20),
    organization VARCHAR(255) DEFAULT '',
    shift_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
INSERT INTO users (id, email, password_hash, role, name)
VALUES (
           gen_random_uuid(),
           'admin@camp.com',
           '$2a$10$HANPtqutN50sU.cmNqtnLuURw4ySq/WS/taOK2jkPmzBxvsXRyEmO',
           'SUPER_ADMIN',
            'MAX'
       );
-- +goose Down
DROP TABLE IF EXISTS shifts;
DROP TABLE IF EXISTS camp_users;
DROP TABLE IF EXISTS camps;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS participants