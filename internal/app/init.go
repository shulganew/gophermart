package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/jmoiron/sqlx"
	"github.com/shulganew/gophermart/internal/accrual"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/services"
	"github.com/shulganew/gophermart/internal/storage"
	"go.uber.org/zap"
)

func InitApp(ctx context.Context, conf *config.Config, db *sqlx.DB) (*services.CalculationService, *services.UserService, *services.AccrualService, *services.OrderService) {
	// Load storage
	stor, err := storage.NewRepo(ctx, db)
	if err != nil {
		zap.S().Errorln("Error connect to DB from env: ", err)
	}

	calcSrv := services.NewCalcService(stor)
	userSrv := services.NewUserService(stor)
	client := accrual.NewAccrualClient(conf)
	accSrv := services.NewAccrualService(stor, conf, client)
	orderSrv := services.NewOrderService(stor)

	// Run observe status of orderses in Accrual service
	accSrv.Run(ctx)

	zap.S().Infoln("Application init complite")

	return calcSrv, userSrv, accSrv, orderSrv
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
