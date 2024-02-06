package main

import (
	"context"
	"net/http"
	"os"

	"github.com/shulganew/gophermart/internal/api/router"
	"github.com/shulganew/gophermart/internal/app"
	"github.com/shulganew/gophermart/internal/config"
	_ "github.com/shulganew/gophermart/migrations"
	"go.uber.org/zap"
)

func main() {
	app.InitLog()

	ctx, cancel := app.InitContext()
	defer cancel()

	conf := config.InitConfig()

	db, err := app.InitDB(ctx, conf)
	if err != nil {
		db = nil
		zap.S().Errorln("Can't connect to Database!", err)
		panic(err)
	}
	defer func() {
		err := db.Close()
		zap.S().Errorln("Could not close db connection", err)
	}()
	// Init application
	container := app.InitApp(ctx, conf, db)

	// Graceful shotdown
	go func(ctx context.Context) {
		<-ctx.Done()
		zap.S().Infoln("Graceful shutdown...")
		os.Exit(0)
	}(ctx)

	//start web
	if err := http.ListenAndServe(conf.Address, router.RouteMarket(container)); err != nil {
		panic(err)
	}
}
