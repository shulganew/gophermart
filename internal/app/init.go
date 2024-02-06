package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"

	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/ports/storage"
	"github.com/shulganew/gophermart/migrations"
	"go.uber.org/zap"
)

func InitApp(ctx context.Context, conf *config.Config, db *sqlx.DB) (container *Container) {

	// Load storage.
	stor, err := storage.NewRepo(ctx, db)
	if err != nil {
		zap.S().Errorln("Error connect to DB from env: ", err)
	}

	// Create config Container.
	container = NewContainer(conf, stor)

	// Run observe status of orderses in Accrual service.
	accSrv := container.GetAccrualService()
	accSrv.Run(ctx)

	zap.S().Infoln("Application init complite")
	return container
}

// Init context from graceful shutdown. Send to all function for return by syscall.SIGINT, syscall.SIGTERM.
func InitContext() (ctx context.Context, cancel context.CancelFunc) {
	exit := make(chan os.Signal, 1)
	ctx, cancel = context.WithCancel(context.Background())
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	go func() {
		<-exit
		cancel()
	}()
	return
}

func InitLog() zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	zap.ReplaceGlobals(logger)
	defer func() {
		_ = logger.Sync()
	}()

	sugar := *logger.Sugar()

	defer func() {
		_ = sugar.Sync()
	}()
	return sugar
}

// Init Database.
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
	// Connection for Gophermart.
	db, err = sqlx.Connect(config.DataBaseType, conf.DSN)
	if err != nil {
		return nil, err
	}

	// Create tables for Market if not exist.
	_, err = db.ExecContext(ctx, migrations.CreateENUM)
	if err != nil {
		return nil, fmt.Errorf("error create processing enum:  %w", err)
	}

	_, err = db.ExecContext(ctx, migrations.CreateUser)
	if err != nil {
		return nil, fmt.Errorf("error create users %w", err)
	}

	_, err = db.ExecContext(ctx, migrations.CreateOrders)
	if err != nil {
		return nil, fmt.Errorf("error create orders %w", err)
	}

	return
}
