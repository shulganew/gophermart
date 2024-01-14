package main

import (
	"net/http"

	"github.com/shulganew/gophermart/internal/api/router"
	"github.com/shulganew/gophermart/internal/app"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/storage"
	"go.uber.org/zap"
)

func main() {

	app.InitLog()

	ctx, cancel := app.InitContext()
	defer cancel()

	conf := config.InitConfig()

	db, err := storage.InitDB(ctx, conf.DSN)
	if err != nil {
		db = nil
		zap.S().Errorln("Can't connect to Database!", err)
		panic(err)
	}
	defer db.Close(ctx)

	//Init application
	market, register := app.InitApp(ctx, *conf, db)

	//start web
	if err := http.ListenAndServe(conf.Address, router.RouteShear(conf, market, register)); err != nil {
		panic(err)
	}
}
