package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/services"
)

// Chi Router for application
func RouteShear(conf *config.Config, market *services.Market, register *services.Register) (r *chi.Mux) {

	r = chi.NewRouter()

	r.Route("/", func(r chi.Router) {

		//api
		//apiHand := handlers.NewHandlerAPI(conf)
		//r.Post("/api/user/register", http.HandlerFunc(apiHand.GetBrief))

	})

	return
}
