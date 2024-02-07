package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/jmoiron/sqlx"

	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/ports/storage"
	"go.uber.org/zap"
)

func InitApp(ctx context.Context) (application *Application, err error) {
	// Get application config.
	conf := config.InitConfig()

	// Connection for Gophermart.
	db, err := sqlx.Connect(config.DataBaseType, conf.DSN)
	if err != nil {
		return nil, err
	}

	// Load storage.
	stor, err := storage.NewRepo(ctx, db)
	if err != nil {
		return nil, err
	}

	// Create config Container
	application = NewApp(conf, stor)

	// Run observe status of orderses in Accrual service.
	accSrv := application.GetAccrualService()
	accSrv.Run(ctx)

	zap.S().Infoln("Application init complite")
	return application, nil
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
