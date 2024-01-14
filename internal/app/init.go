package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/services"
	"github.com/shulganew/gophermart/internal/storage"
	"go.uber.org/zap"
)

func InitApp(ctx context.Context, conf config.Config, db *pgx.Conn) (*services.Market, *services.Register) {

	// Load storage
	stor, err := storage.NewRepo(ctx, db)
	if err != nil {
		zap.S().Errorln("Error connect to DB from env: ", err)

	}

	market := services.NewMarket(stor)

	register := services.NewRegister()

	zap.S().Infoln("Application init complite")

	return market, register

}

// Init context from graceful shutdown. Send to all function for return by syscall.SIGINT, syscall.SIGTERM
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
	defer logger.Sync()

	sugar := *logger.Sugar()

	defer sugar.Sync()
	return sugar
}
