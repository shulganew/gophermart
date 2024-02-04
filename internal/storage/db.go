package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"

	"github.com/shulganew/gophermart/db/tables"
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
	_, err = db.ExecContext(ctx, tables.CreateENUM)
	if err != nil {
		return nil, fmt.Errorf("error create processing enum:  %w", err)
	}

	_, err = db.ExecContext(ctx, tables.CreateUser)
	if err != nil {
		return nil, fmt.Errorf("error create users %w", err)
	}

	_, err = db.ExecContext(ctx, tables.CreateOrders)
	if err != nil {
		return nil, fmt.Errorf("error create orders %w", err)
	}

	return
}

func (base *Repo) Start(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	err := base.master.PingContext(ctx)
	defer cancel()
	return err
}
