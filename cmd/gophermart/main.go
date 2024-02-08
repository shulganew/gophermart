package main

import (
	"context"
	"os"

	"github.com/shulganew/gophermart/internal/app"
	"github.com/shulganew/gophermart/internal/app/server"
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

	// Close DB connection.
	defer func() {
		err := application.Repo().DB().Close()
		zap.S().Errorln("Could not close db connection", err)
	}()

	// Graceful shotdown.
	go func(ctx context.Context) {
		<-ctx.Done()
		zap.S().Infoln("Graceful shutdown...")
		os.Exit(0)
	}(ctx)

	// Run server.
	server.NewMarket(application).Run()
}
