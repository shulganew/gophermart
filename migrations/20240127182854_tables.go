package migrations

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/shulganew/gophermart/internal/storage"
)

func init() {
	goose.AddMigrationContext(upTables, downTables)
}

func upTables(ctx context.Context, tx *sql.Tx) error {
	s := make([]string, 0)
	s = append(s, storage.CreateENUM)
	s = append(s, storage.CreateUser)
	s = append(s, storage.CreateOrders)

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
