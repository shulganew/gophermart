package router

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shulganew/gophermart/internal/api/handlers"
	"github.com/shulganew/gophermart/internal/api/middlewares"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/services"
)

// Chi Router for application
func RouteMarket(conf *config.Config, market *services.Market, register *services.Register, observer *services.Observer) (r *chi.Mux) {

	r = chi.NewRouter()

	r.Route("/", func(r chi.Router) {

		//send password for enctription to middlewares
		r.Use(func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				ctx := context.WithValue(r.Context(), config.CtxPassKey{}, conf.PassJWT)
				h.ServeHTTP(w, r.WithContext(ctx))
			})
		})

		userReg := handlers.NewHandlerRegister(conf, register)
		r.Post("/api/user/register", http.HandlerFunc(userReg.SetUser))

		userLogin := handlers.NewHandlerLogin(conf, register)
		r.Post("/api/user/login", http.HandlerFunc(userLogin.LoginUser))

		r.Route("/api/user", func(r chi.Router) {
			r.Use(middlewares.Auth)
			orders := handlers.NewHandlerOrder(conf, market, observer)
			r.Post("/orders", http.HandlerFunc(orders.SetOrder))
			r.Get("/orders", http.HandlerFunc(orders.GetOrders))

			balance := handlers.NewHandlerBalance(conf, market)
			r.Get("/balance", http.HandlerFunc(balance.GetBalance))
			r.Post("/balance/withdraw", http.HandlerFunc(balance.WithdrawnBalance))
		})

	})

	return
}
