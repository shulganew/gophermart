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
    uploaded TIMESTAMPTZ NOT NULL,
    status processing NOT NULL DEFAULT 'NEW'
    );

CREATE TABLE bonuses (
    id SERIAL, 
    onumber VARCHAR(20) NOT NULL REFERENCES orders(onumber) UNIQUE,
    bonus_used NUMERIC DEFAULT 0,
    bonus_accrual NUMERIC DEFAULT 0
);
	
    
/*
SELECT SUM(bonuses.bonus_accrual)
		FROM orders 
		INNER JOIN users ON orders.user_id = users.user_id
		INNER JOIN bonuses ON orders.onumber = bonuses.onumber;


SELECT SUM(bonuses.bonus_used)
		FROM orders 
		INNER JOIN users ON orders.user_id = users.user_id
		INNER JOIN bonuses ON orders.onumber = bonuses.onumber;
		
SELECT users.user_id, orders.onumber, orders.uploaded, orders.status, bonuses.bonus_used, bonuses.bonus_accrual
		FROM orders 
		INNER JOIN users ON orders.user_id = users.user_id
		INNER JOIN bonuses ON orders.onumber = bonuses.onumber;

   UPDATE orders SET status = 'PROCESSED' WHERE onumber = '1816868606';
   UPDATE bonuses SET  bonus_accrual = '10' WHERE onumber = '1816868606';
/*

