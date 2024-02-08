-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'processing') THEN
		CREATE TYPE processing AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED', 'REGISTERED');
	END IF;
END$$;

CREATE TABLE IF NOT EXISTS users (
	id SERIAL, 
	user_id UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(), 
	login TEXT NOT NULL UNIQUE, 
	password_hash TEXT NOT NULL,
	withdrawals NUMERIC DEFAULT 0,
	bonuses NUMERIC DEFAULT 0
	);

CREATE TABLE IF NOT EXISTS users (
	id SERIAL, 
	user_id UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(), 
	login TEXT NOT NULL UNIQUE, 
	password_hash TEXT NOT NULL,
	withdrawals NUMERIC DEFAULT 0,
	bonuses NUMERIC DEFAULT 0
	);

CREATE TABLE IF NOT EXISTS orders (
		id SERIAL, 
		user_id UUID NOT NULL REFERENCES users(user_id),
		order_number VARCHAR(20) NOT NULL UNIQUE,
		is_preorder BOOLEAN NOT NULL, 
		uploaded TIMESTAMPTZ NOT NULL,
		status processing NOT NULL DEFAULT 'NEW',
		withdrawn NUMERIC DEFAULT 0,
		accrual NUMERIC DEFAULT 0
		);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE orders;
DROP TABLE users;
-- +goose StatementEnd