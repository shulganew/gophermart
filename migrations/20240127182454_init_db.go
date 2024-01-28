package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {

	goose.AddMigrationNoTxContext(upInitDb, downInitDb)
}

func upInitDb(ctx context.Context, db *sql.DB) error {

	s := make([]string, 0)
	s = append(s, "CREATE USER market WITH ENCRYPTED PASSWORD '1'")
	s = append(s, "CREATE USER praktikum WITH ENCRYPTED PASSWORD 'praktikum'")
	s = append(s, "CREATE DATABASE market")
	s = append(s, "CREATE DATABASE praktikum")
	s = append(s, "GRANT ALL PRIVILEGES ON DATABASE market TO market")
	s = append(s, "GRANT ALL PRIVILEGES ON DATABASE praktikum TO market")
	s = append(s, "ALTER DATABASE market OWNER TO market")
	s = append(s, "ALTER DATABASE praktikum OWNER TO market")

	for _, q := range s {
		_, err := db.ExecContext(ctx, q)
		if err != nil {
			return err
		}
	}

	return nil
}

func downInitDb(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, "DROP DATABASE market")
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, "DROP DATABASE praktikum")
	if err != nil {
		return err
	}

	return nil
}
