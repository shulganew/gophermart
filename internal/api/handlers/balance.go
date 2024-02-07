package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/shopspring/decimal"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/entities"
	"github.com/shulganew/gophermart/internal/services"
	"go.uber.org/zap"
)

type HandlerBalance struct {
	calcSrv  *services.CalculationService
	conf     *config.Config
	orderSrv *services.OrderService
}

func NewHandlerBalance(conf *config.Config, calc *services.CalculationService, orders *services.OrderService) *HandlerBalance {
	return &HandlerBalance{calcSrv: calc, conf: conf, orderSrv: orders}
}

func (u *HandlerBalance) GetBalance(res http.ResponseWriter, req *http.Request) {
	// get UserID from cxt values
	ctxConfigVal := req.Context().Value(entities.MiddlwDTO{})
	ctxConfig, ok := ctxConfigVal.(entities.MiddlwDTO)
	if !ok {
		errt := "Cat't get MiddlwDTO from context."
		zap.S().Errorln(errt)
		http.Error(res, errt, http.StatusInternalServerError)
		return
	}

	// Check from middleware is user authorized 401
	if !ctxConfig.IsRegistered() {
		http.Error(res, "JWT not found.", http.StatusUnauthorized)
	}

	userID := ctxConfig.GetUserID()

	bonuses, err := u.calcSrv.GetBonuses(req.Context(), userID)
	if err != nil {
		// 500
		errt := "Cat't get bonuses."
		zap.S().Errorln(errt, err)
		http.Error(res, errt, http.StatusInternalServerError)
		return
	}

	withdrawn, err := u.calcSrv.GetWithdrawn(req.Context(), userID)
	if err != nil {
		// 500
		errt := "Cat't get withdrawn."
		zap.S().Errorln(errt, err)
		http.Error(res, errt, http.StatusInternalServerError)
		return
	}

	userBalance := entities.NewUserBalance(bonuses, withdrawn)

	jsonBalance, err := json.Marshal(userBalance)
	if err != nil {
		http.Error(res, "Error during Marshal user's balance", http.StatusInternalServerError)
	}

	// set content type
	res.Header().Add("Content-Type", "application/json")

	// set status code 200
	res.WriteHeader(http.StatusOK)

	_, err = res.Write([]byte(jsonBalance))
	if err != nil {
		zap.S().Errorln("Can't write to response in get balance handler", err)
	}
}

func (u *HandlerBalance) SetWithdraw(res http.ResponseWriter, req *http.Request) {
	// get UserID from cxt values
	ctxConfigVal := req.Context().Value(entities.MiddlwDTO{})
	ctxConfig, ok := ctxConfigVal.(entities.MiddlwDTO)
	if !ok {
		errt := "Cat't get MiddlwDTO from context."
		zap.S().Errorln(errt)
		http.Error(res, errt, http.StatusInternalServerError)
		return
	}

	// Check from middleware is user authorized 401
	if !ctxConfig.IsRegistered() {
		http.Error(res, "JWT not found.", http.StatusUnauthorized)
	}

	userID := ctxConfig.GetUserID()

	var wd entities.Withdraw
	if err := json.NewDecoder(req.Body).Decode(&wd); err != nil {
		// If can't decode 400
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	zap.S().Debugln("Withdrawn order: ", wd.OrderNr)
	zap.S().Debugln("Withdrawn amount: ", wd.Withdrawn)
	err := goluhn.Validate(wd.OrderNr)
	if err != nil {
		// 422
		errt := "Order not luna valid."
		zap.S().Debugln(errt, wd.OrderNr)
		http.Error(res, errt, http.StatusUnprocessableEntity)
		return
	}

	amount := decimal.NewFromFloat(wd.Withdrawn)

	isEnough, err := u.calcSrv.CheckBalance(req.Context(), userID, amount)
	if err != nil {
		// 500
		errt := "Error cheking bonuses balance."
		zap.S().Errorln(errt, wd.OrderNr, err)
		http.Error(res, errt, http.StatusInternalServerError)
		return
	}
	if !isEnough {
		// 402
		http.Error(res, "Not enuogh bonuses.", http.StatusPaymentRequired)
		return
	}

	// Create preorder with withdrawal and add storage with mark preoreder bool = true
	order := entities.NewOrder(userID, wd.OrderNr, true, amount, decimal.Zero)
	existed, err := u.orderSrv.IsExist(req.Context(), order.OrderNr)
	if err != nil {
		// 500
		errt := "Error during withdraw. Checking order duplication error."
		zap.S().Debugln(errt, wd.OrderNr, err)
		http.Error(res, errt, http.StatusInternalServerError)
		return
	}
	if existed {
		// 422
		errt := "Order alredy existed."
		zap.S().Debugln(errt, wd.OrderNr)
		http.Error(res, errt, http.StatusUnprocessableEntity)
		return
	}

	err = u.orderSrv.AddOrder(req.Context(), true, order)
	if err != nil {
		// 500
		errt := "Error during withdraw. Adding preorder error."
		zap.S().Debugln(errt, wd.OrderNr, err)
		http.Error(res, errt, http.StatusInternalServerError)
		return
	}

	// Update withdrawals and bonuses balance.
	err = u.calcSrv.MakeWithdrawn(req.Context(), userID, amount)
	if err != nil {
		// 500
		errt := "Error during withdrawn."
		zap.S().Debugln(errt, wd.OrderNr, err)
		http.Error(res, "Error during withdraw. Update balbance error", http.StatusInternalServerError)
		return
	}
	// set status code 200
	res.WriteHeader(http.StatusOK)

	_, err = res.Write([]byte("Done."))
	if err != nil {
		zap.S().Errorln("Can't write to response in SetWithdrawn  handler", err)
	}
}

func (u *HandlerBalance) GetWithdrawals(res http.ResponseWriter, req *http.Request) {
	// get UserID from cxt values
	ctxConfigVal := req.Context().Value(entities.MiddlwDTO{})
	ctxConfig, ok := ctxConfigVal.(entities.MiddlwDTO)
	if !ok {
		errt := "Cat't get MiddlwDTO from context."
		zap.S().Errorln(errt)
		http.Error(res, errt, http.StatusInternalServerError)
		return
	}

	// Check from middleware is user authorized 401
	if !ctxConfig.IsRegistered() {
		http.Error(res, "JWT not found.", http.StatusUnauthorized)
	}

	userID := ctxConfig.GetUserID()

	withdrawals, err := u.calcSrv.GetWithdrawals(req.Context(), userID)
	if err != nil {
		// 500
		errt := "Cat't get withdrawals"
		zap.S().Error(errt, err)
		http.Error(res, errt, http.StatusInternalServerError)
		return
	}
	if len(withdrawals) == 0 {
		// 204 - no withdrawals
		http.Error(res, "Cat't get withdrawals", http.StatusNoContent)
		return
	}

	jsonWithdraw, err := json.Marshal(withdrawals)
	if err != nil {
		errt := "Error during Marshal user's balance"
		zap.S().Error(errt, err)
		http.Error(res, errt, http.StatusInternalServerError)
	}

	// set content type
	res.Header().Add("Content-Type", "application/json")

	// set status code 200
	res.WriteHeader(http.StatusOK)

	_, err = res.Write([]byte(jsonWithdraw))
	if err != nil {
		zap.S().Errorln("Can't write to response in GetWithdrawals  handler", err)
	}
}
