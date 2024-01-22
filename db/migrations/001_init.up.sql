CREATE TYPE processing AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED', 'REGISTERED');

CREATE TABLE users (
    id SERIAL, 
    user_id UUID NOT NULL UNIQUE, 
    login TEXT NOT NULL UNIQUE, 
    password TEXT NOT NULL
    );



CREATE TABLE orders (
    id SERIAL, 
    user_id UUID NOT NULL REFERENCES users(user_id),
    onumber VARCHAR(20) NOT NULL UNIQUE,
    uploaded TIMESTAMPTZ NOT NULL,
    isPreorder BOOLEAN NOT NULL,
    status processing NOT NULL DEFAULT 'NEW'
    );

CREATE TABLE bonuses (
    id SERIAL, 
    onumber VARCHAR(20) NOT NULL REFERENCES orders(onumber) UNIQUE,
    bonus_used NUMERIC DEFAULT 0,
    bonus_accrual NUMERIC DEFAULT 0
);

