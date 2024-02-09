package server

import (
	"net/http"

	"github.com/shulganew/gophermart/internal/api/router"
	"github.com/shulganew/gophermart/internal/app"
)

type Market struct {
	application *app.UseCases
}

func NewMarket(appl *app.UseCases) *Market {
	return &Market{application: appl}
}

func (s Market) Run() {
	// Start web server.
	if err := http.ListenAndServe(s.application.Config().Address, router.RouteMarket(s.application)); err != nil {
		panic(err)
	}
}
