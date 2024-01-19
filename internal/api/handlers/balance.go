package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/shopspring/decimal"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/services"
)

type UserBalance struct {
	Bonus     float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func NewUserBalance(bonus *decimal.Decimal, withdrawn *decimal.Decimal) *UserBalance {

	b := bonus.InexactFloat64()
	w := withdrawn.InexactFloat64()
	return &UserBalance{Bonus: b, Withdrawn: w}
}

type HandlerBalance struct {
	market *services.Market
	conf   *config.Config
}

func NewHandlerBalance(conf *config.Config, market *services.Market) *HandlerBalance {

	return &HandlerBalance{market: market, conf: conf}
}

func (u *HandlerBalance) GetBalance(res http.ResponseWriter, req *http.Request) {
	// get UserID from cxt values
	ctxConfig := req.Context().Value(config.CtxConfig{}).(config.CtxConfig)

	// Check from middleware is user authorized 401
	if !ctxConfig.IsRegistered() {
		http.Error(res, "JWT not found.", http.StatusUnauthorized)
	}

	userID := ctxConfig.GetUserID()

	acc, withdrawn, err := u.market.GetBalance(req.Context(), userID)
	if err != nil {
		// 500
		http.Error(res, "Cat't get orders", http.StatusInternalServerError)
		return
	}

	balance := acc.Sub(*withdrawn)

	userBalance := NewUserBalance(&balance, withdrawn)

	jsonBalance, err := json.Marshal(userBalance)
	if err != nil {
		http.Error(res, "Error during Marshal user's balance", http.StatusInternalServerError)
	}

	//set status code 200
	res.WriteHeader(http.StatusOK)

	res.Write([]byte(jsonBalance))

}
