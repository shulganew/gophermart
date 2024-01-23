package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/shopspring/decimal"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/model"
	"github.com/shulganew/gophermart/internal/services"
	"go.uber.org/zap"
)

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

	userBalance := model.NewUserBalance(&balance, withdrawn)

	jsonBalance, err := json.Marshal(userBalance)
	if err != nil {
		http.Error(res, "Error during Marshal user's balance", http.StatusInternalServerError)
	}

	//set content type
	res.Header().Add("Content-Type", "application/json")

	//set status code 200
	res.WriteHeader(http.StatusOK)

	res.Write([]byte(jsonBalance))

}

func (u *HandlerBalance) SetWithdraw(res http.ResponseWriter, req *http.Request) {
	// get UserID from cxt values
	ctxConfig := req.Context().Value(config.CtxConfig{}).(config.CtxConfig)

	// Check from middleware is user authorized 401
	if !ctxConfig.IsRegistered() {
		http.Error(res, "JWT not found.", http.StatusUnauthorized)
	}

	userID := ctxConfig.GetUserID()

	var wd model.Withdraw
	if err := json.NewDecoder(req.Body).Decode(&wd); err != nil {
		// If can't decode 400
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	zap.S().Debugln("Withdrawn order: ", wd.Onumber)
	zap.S().Debugln("Withdrawn amount: ", wd.Withdrawn)
	err := goluhn.Validate(wd.Onumber)
	if err != nil {
		// 422
		zap.S().Debugln("Order not luna valid: ", wd.Onumber)
		http.Error(res, "Order nuber not vaild.", http.StatusUnprocessableEntity)
		return
	}

	amount := decimal.NewFromFloat(wd.Withdrawn)

	isEnough, err := u.market.CheckBalance(req.Context(), userID, &amount)
	if err != nil {
		// 500
		http.Error(res, "Error cheking bonuses balance.", http.StatusInternalServerError)
		return
	}
	if !isEnough {
		// 402
		http.Error(res, "Not enuogh bonuses.", http.StatusPaymentRequired)
		return
	}

	order := model.NewOrder(userID, wd.Onumber, true, &amount, &decimal.Zero)
	existed, err := u.market.SetOrder(req.Context(), true, order)
	if existed {
		// 422
		errt := "Order alredy existed."
		zap.S().Debugln(errt, wd.Onumber)
		http.Error(res, errt, http.StatusUnprocessableEntity)
		return
	}
	if err != nil {
		// 500
		http.Error(res, "Error during withdraw.", http.StatusInternalServerError)
		return
	}
	//set status code 200
	res.WriteHeader(http.StatusOK)

	res.Write([]byte("Done."))

}

func (u *HandlerBalance) GetWithdrawals(res http.ResponseWriter, req *http.Request) {

	// get UserID from cxt values
	ctxConfig := req.Context().Value(config.CtxConfig{}).(config.CtxConfig)

	// Check from middleware is user authorized 401
	if !ctxConfig.IsRegistered() {
		http.Error(res, "JWT not found.", http.StatusUnauthorized)
	}

	userID := ctxConfig.GetUserID()

	withdrawals, err := u.market.GetWithdrawals(req.Context(), userID)
	if err != nil {
		// 500
		zap.S().Error("Cat't get withdrawals", err)
		http.Error(res, "Cat't get withdrawals", http.StatusInternalServerError)
		return
	}
	if len(withdrawals) == 0 {
		// 204 - no withdrawals
		http.Error(res, "Cat't get withdrawals", http.StatusNoContent)
		return
	}

	jsonWithdraw, err := json.Marshal(withdrawals)
	if err != nil {
		http.Error(res, "Error during Marshal user's balance", http.StatusInternalServerError)
	}

	//set content type
	res.Header().Add("Content-Type", "application/json")

	//set status code 200
	res.WriteHeader(http.StatusOK)

	res.Write([]byte(jsonWithdraw))

}
