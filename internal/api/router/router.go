package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shulganew/gophermart/internal/api/handlers"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/services"
)

// Chi Router for application
func RouteShear(conf *config.Config, market *services.Market, register *services.Register) (r *chi.Mux) {

	r = chi.NewRouter()

	r.Route("/", func(r chi.Router) {

		userReg := handlers.NewHandlerRegister(conf, register)
		r.Post("/api/user/register", http.HandlerFunc(userReg.SetUser))

		userLogin := handlers.NewHandlerLogin(conf, register)
		r.Post("/api/user/login", http.HandlerFunc(userLogin.LoginUser))
	})

	return
}
