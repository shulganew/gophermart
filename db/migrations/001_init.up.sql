CREATE TABLE users (
    id SERIAL, 
    user_id UUID NOT NULL UNIQUE, 
    login TEXT NOT NULL UNIQUE, 
    password TEXT NOT NULL
    );

CREATE TYPE processing AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED', 'REGISTERED');

CREATE TABLE orders (
    id SERIAL, 
    user_id UUID NOT NULL REFERENCES users(user_id),
    onumber VARCHAR(20) NOT NULL UNIQUE,
    uploaded TIMESTAMP NOT NULL,
    status processing NOT NULL DEFAULT 'NEW'
    );

CREATE TABLE bonuses (
    id SERIAL, 
    onumber VARCHAR(20) NOT NULL REFERENCES orders(onumber),
    bonus_used INT DEFAULT 0,
    bonus_accural INT DEFAULT 0,
    accrual_time TIMESTAMP NOT NULL
);



