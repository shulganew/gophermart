package migrations

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upTables, downTables)
}

func upTables(ctx context.Context, tx *sql.Tx) error {
	s := make([]string, 0)
	s = append(s, "CREATE TYPE processing AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED', 'REGISTERED');")
	s = append(s, `CREATE TABLE users (
		id SERIAL, 
		user_id UUID NOT NULL UNIQUE, 
		login TEXT NOT NULL UNIQUE, 
		password TEXT NOT NULL
		);`)

	s = append(s, `CREATE TABLE orders (
		id SERIAL, 
		user_id UUID NOT NULL REFERENCES users(user_id),
		onumber VARCHAR(20) NOT NULL UNIQUE,
		uploaded TIMESTAMPTZ NOT NULL,
		is_preorder BOOLEAN NOT NULL,
		status processing NOT NULL DEFAULT 'NEW'
		);`)

	s = append(s, `CREATE TABLE bonuses (
		id SERIAL, 
		onumber VARCHAR(20) NOT NULL REFERENCES orders(onumber) UNIQUE,
		bonus_used NUMERIC DEFAULT 0,
		bonus_accrual NUMERIC DEFAULT 0
	);`)

	for _, q := range s {
		_, err := tx.ExecContext(ctx, q)
		if err != nil {
			return err
		}
	}
	return nil
}

func downTables(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "DROP TABLE bonuses")
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "DROP TABLE orders")
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "DROP TABLE users")
	if err != nil {
		return err
	}
	return nil
}
