package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ShiraazMoollatjie/goluhn"
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

type Withdraw struct {
	Onumber   string  `json:"order"`
	Withdrawn float64 `json:"sum"`
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

func (u *HandlerBalance) WithdrawnBalance(res http.ResponseWriter, req *http.Request) {
	// get UserID from cxt values
	ctxConfig := req.Context().Value(config.CtxConfig{}).(config.CtxConfig)

	// Check from middleware is user authorized 401
	if !ctxConfig.IsRegistered() {
		http.Error(res, "JWT not found.", http.StatusUnauthorized)
	}

	userID := ctxConfig.GetUserID()

	var wd Withdraw
	if err := json.NewDecoder(req.Body).Decode(&wd); err != nil {
		// If can't decode 400
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	err := goluhn.Validate(wd.Onumber)
	if err != nil {
		// 422
		http.Error(res, "Order nuber not vaild.", http.StatusUnprocessableEntity)
		return
	}

	amount := decimal.NewFromFloat(wd.Withdrawn)

	existed, err := u.market.IsExistForUser(req.Context(), userID, wd.Onumber)
	if err != nil {
		// 500
		http.Error(res, "Error cheking order for user.", http.StatusInternalServerError)
		return
	}
	if !existed {
		// 422
		http.Error(res, "Order not existed.", http.StatusUnprocessableEntity)
		return
	}

	isEnough, err := u.market.CheckBalance(req.Context(), userID, &amount)
	if err != nil {
		// 500
		http.Error(res, "Error cheking bonuses balance.", http.StatusInternalServerError)
		return
	}
	if !isEnough {
		// 422
		http.Error(res, "Not enuogh bonuses.", http.StatusPaymentRequired)
		return
	}

	err = u.market.Withdrow(req.Context(), userID, wd.Onumber, &amount)
	if err != nil {
		// 500
		http.Error(res, "Error during withdraw.", http.StatusInternalServerError)
		return
	}
	//set status code 200
	res.WriteHeader(http.StatusOK)

	res.Write([]byte("Done."))

}
