package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/shulganew/gophermart/internal/config"

	"go.uber.org/zap"
)

type Repo struct {
	master *sqlx.DB
}

func NewRepo(ctx context.Context, master *sqlx.DB) (*Repo, error) {
	db := Repo{master: master}
	err := db.Start(ctx)
	return &db, err
}

// Init Database
func InitDB(ctx context.Context, conf *config.Config) (db *sqlx.DB, err error) {

	// Migrations enebles in config
	if conf.Migrations {
		zap.S().Infoln("Migrations is start:")
		//Init connection for admin user for prepare databse and make migrations
		initdb, err := goose.OpenDBWithDriver(config.DataBaseType, conf.DSNMitration)
		if err != nil {
			zap.S().Fatalln("goose: failed to open DB: %v\n", err)
		}

		defer func() {
			if err := initdb.Close(); err != nil {
				zap.S().Fatalln("goose: failed to close DB: %v\n", err)
			}
		}()

		//Init database migrations
		if err := goose.UpContext(ctx, initdb, "migrations"); err != nil { //
			zap.S().Fatalln("Error make databes migrations before starting Market app: ", err)
		} else {
			zap.S().Infoln("Migrations update...")
		}
	}
	//Connection for Gophermart
	db, err = sqlx.Connect(config.DataBaseType, conf.DSN)
	if err != nil {
		return nil, err
	}

	// Create tables for Market if not exist
	query := `
	DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'processing') THEN
			CREATE TYPE processing AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED', 'REGISTERED');
		END IF;
	END$$
	`
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error create processing enum:  %w", err)
	}

	query = `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL, 
		user_id UUID NOT NULL UNIQUE, 
		login TEXT NOT NULL UNIQUE, 
		password TEXT NOT NULL
		);
		`

	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error create users %w", err)
	}

	query = `
	CREATE TABLE IF NOT EXISTS orders (
		id SERIAL, 
		user_id UUID NOT NULL REFERENCES users(user_id),
		onumber VARCHAR(20) NOT NULL UNIQUE,
		is_preorder BOOLEAN NOT NULL, 
		uploaded TIMESTAMPTZ NOT NULL,
		status processing NOT NULL DEFAULT 'NEW'
		);
		`
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error create orders %w", err)
	}

	query = `
	CREATE TABLE IF NOT EXISTS bonuses (
		id SERIAL, 
		onumber VARCHAR(20) NOT NULL REFERENCES orders(onumber) UNIQUE,
		bonus_used NUMERIC DEFAULT 0,
		bonus_accrual NUMERIC DEFAULT 0
	);
	`
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error create bonuses %w", err)
	}

	return
}

func (base *Repo) Start(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	err := base.master.PingContext(ctx)
	defer cancel()
	return err
}
