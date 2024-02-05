package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/shopspring/decimal"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/model"
	"github.com/shulganew/gophermart/internal/services"
	"go.uber.org/zap"
)

type HandlerOrder struct {
	calcSrv  *services.CalculationService
	conf     *config.Config
	accSrv   *services.AccrualService
	orderSrv *services.OrderService
}

func NewHandlerOrder(conf *config.Config, calc *services.CalculationService, accSrv *services.AccrualService, orderSrv *services.OrderService) *HandlerOrder {

	return &HandlerOrder{calcSrv: calc, conf: conf, accSrv: accSrv, orderSrv: orderSrv}
}

func (u *HandlerOrder) AddOrder(res http.ResponseWriter, req *http.Request) {

	//get UserID from cxt values
	ctxConfig := req.Context().Value(model.MiddlwDTO{}).(model.MiddlwDTO)

	// Check from middleware is user authorized
	if !ctxConfig.IsRegistered() {
		http.Error(res, "JWT not found.", http.StatusUnauthorized)
		return
	}

	userID := ctxConfig.GetUserID()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		// 400
		http.Error(res, "Cat't read body data", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	orderNr := string(body)
	zap.S().Infoln("Set Order for user: ", userID, " Order: ", orderNr)
	// Create order
	order := model.NewOrder(userID, orderNr, false, decimal.Zero, decimal.Zero)

	isValid := order.IsValid()
	if !isValid {
		// 422
		errt := "Order nuber not vaild."
		zap.S().Debugln(errt, orderNr)
		http.Error(res, errt, http.StatusUnprocessableEntity)
		return
	}

	isExist, err := u.orderSrv.IsExist(req.Context(), order.OrderNr)
	if err != nil {
		// 500
		errt := "Get error during checkig if order existed."
		zap.S().Error(errt, orderNr, err)
		http.Error(res, errt, http.StatusInternalServerError)
		return
	}

	if !isExist {
		err = u.orderSrv.AddOrder(req.Context(), false, order)
		if err != nil {
			// 500
			errt := "Get error during save new order."
			zap.S().Error(errt, orderNr, err)
			http.Error(res, errt, http.StatusInternalServerError)
			return
		}

		zap.S().Infoln("New order added: ", order.OrderNr)
		// 202 - New order
		res.WriteHeader(http.StatusAccepted)

		res.Write([]byte("Set order!" + orderNr))
		return

	}

	//Order already existed
	//Check if order created befower with withdraw as prepaid
	isPreorder, err := u.calcSrv.IsPreOrder(req.Context(), userID, orderNr)
	if err != nil {
		errt := "Get error during preorder search."
		zap.S().Error(errt, err, orderNr)
		http.Error(res, errt, http.StatusInternalServerError)
		return
	}

	if isPreorder {
		// Move prepaid preoreder to regular order.
		err = u.calcSrv.MovePreOrder(req.Context(), order)
		if err != nil {
			errt := "Get error during preorder update."
			zap.S().Debugln(errt, orderNr)
			http.Error(res, errt, http.StatusInternalServerError)
			return
		}

	}

	// Is Existed for this user.
	isExistUser, err := u.orderSrv.IsExistForUser(req.Context(), userID, orderNr)
	if err != nil {
		errt := "Get error during search duplicated order for user."
		zap.S().Error(errt, orderNr)
		http.Error(res, errt, http.StatusInternalServerError)
		return
	}
	if isExistUser {
		// 200 alredy created for user
		errt := "Order duplicated for User."

		zap.S().Debugln(errt, orderNr)
		http.Error(res, errt, http.StatusOK)
		return
	}

	// 409
	errt := "Order duplicated for Other User."
	zap.S().Debugln(errt, orderNr)
	http.Error(res, errt, http.StatusConflict)

}

func (u *HandlerOrder) GetOrders(res http.ResponseWriter, req *http.Request) {

	//get UserID from cxt values
	ctxConfig := req.Context().Value(model.MiddlwDTO{}).(model.MiddlwDTO)

	// Check from middleware is user authorized 401
	if !ctxConfig.IsRegistered() {
		errt := "JWT not found."
		zap.S().Infoln(errt)
		http.Error(res, errt, http.StatusUnauthorized)
		return
	}

	userID := ctxConfig.GetUserID()

	// Load user's orders
	orders, err := u.orderSrv.GetOrders(req.Context(), userID)

	zap.S().Infoln("GetOrders len", len(orders), "for user: ", userID)
	for _, ord := range orders {
		zap.S().Infoln("Orders:", ord)
	}

	if err != nil {
		// 500
		errt := "Cat't get orders:"
		zap.S().Errorln(errt, err)
		http.Error(res, errt, http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		// 204
		errt := "No content"
		zap.S().Debugln(errt)
		http.Error(res, errt, http.StatusNoContent)
		return
	}

	jsonOrders, err := json.Marshal(orders)
	if err != nil {
		errt := "Error during Marshal answer Orders"
		zap.S().Errorln(errt, err)
		http.Error(res, errt, http.StatusInternalServerError)
		return
	}

	zap.S().Infoln("Get Orders: ", string(jsonOrders))

	//set content type
	res.Header().Add("Content-Type", "application/json")

	//set status code 200
	res.WriteHeader(http.StatusOK)

	res.Write([]byte(jsonOrders))

}
