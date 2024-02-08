package main

import (
	"context"
	"net/http"
	"os"

	"github.com/shulganew/gophermart/internal/api/router"
	"github.com/shulganew/gophermart/internal/app"
	"go.uber.org/zap"
)

func main() {
	// Init application logging.
	app.InitLog()

	// Init application context.
	ctx, cancel := app.InitContext()
	defer cancel()

	// Init application.
	application, err := app.InitApp(ctx)
	if err != nil {
		panic(err)
	}

	// Close DB connection
	defer func() {
		err := application.Repo().DB().Close()
		zap.S().Errorln("Could not close db connection", err)
	}()

	// Graceful shotdown
	go func(ctx context.Context) {
		<-ctx.Done()
		zap.S().Infoln("Graceful shutdown...")
		os.Exit(0)
	}(ctx)

	//start web
	if err := http.ListenAndServe(application.Config().Address, router.RouteMarket(application)); err != nil {
		panic(err)
	}
}
