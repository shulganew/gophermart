package router

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shulganew/gophermart/internal/api/handlers"
	"github.com/shulganew/gophermart/internal/api/middlewares"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/model"
	"github.com/shulganew/gophermart/internal/services"
)

// Chi Router for application.
func RouteMarket(conf *config.Config, calcSrv *services.CalculationService, userSrv *services.UserService, accSrv *services.AccrualService, orderSrv *services.OrderService) (r *chi.Mux) {
	r = chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		// send password for enctription to middlewares
		r.Use(func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), model.CtxPassKey{}, conf.PassJWT)
				h.ServeHTTP(w, r.WithContext(ctx))
			})
		})

		userReg := handlers.NewHandlerRegister(conf, userSrv)
		r.Post("/api/user/register", http.HandlerFunc(userReg.SetUser))

		userLogin := handlers.NewHandlerLogin(conf, userSrv)
		r.Post("/api/user/login", http.HandlerFunc(userLogin.LoginUser))

		r.Route("/api/user", func(r chi.Router) {
			r.Use(middlewares.Auth)
			orderHand := handlers.NewHandlerOrder(conf, calcSrv, accSrv, orderSrv)
			r.Post("/orders", http.HandlerFunc(orderHand.AddOrder))
			r.Get("/orders", http.HandlerFunc(orderHand.GetOrders))

			balance := handlers.NewHandlerBalance(conf, calcSrv, orderSrv)
			r.Get("/balance", http.HandlerFunc(balance.GetBalance))
			r.Post("/balance/withdraw", http.HandlerFunc(balance.SetWithdraw))
			r.Get("/withdrawals", http.HandlerFunc(balance.GetWithdrawals))
		})
	})

	return
}
