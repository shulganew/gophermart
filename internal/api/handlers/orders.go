package handlers

import (
	"io"
	"net/http"

	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/model"
	"github.com/shulganew/gophermart/internal/services"
	"go.uber.org/zap"
)

type HandlerOrder struct {
	market   *services.Market
	conf     *config.Config
	observer *services.Observer
}

func NewHandlerOrder(conf *config.Config, market *services.Market, observer *services.Observer) *HandlerOrder {

	return &HandlerOrder{market: market, conf: conf, observer: observer}
}

func (u *HandlerOrder) SetOrder(res http.ResponseWriter, req *http.Request) {

	//get UserID from cxt values
	ctxConfig := req.Context().Value(config.CtxConfig{}).(config.CtxConfig)

	// Check from middleware is user authorized
	if !ctxConfig.IsRegistered() {
		http.Error(res, "JWT not found.", http.StatusUnauthorized)
		return
	}

	userID := ctxConfig.GetUserID()

	zap.S().Infoln("Set Order for user: ", userID)

	body, err := io.ReadAll(req.Body)
	if err != nil {
		// 400
		http.Error(res, "Cat't read body data", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	onumber := string(body)

	// Create order

	order := model.NewOrder(userID, onumber)

	isValid := order.IsValid()
	if !isValid {
		// 422
		http.Error(res, "Order nuber not vaild.", http.StatusUnprocessableEntity)
		return
	}

	zap.S().Infoln("Order: ", order)

	isExist, err := u.market.SetOrder(req.Context(), order)
	if !isExist && err != nil {
		// 500
		http.Error(res, "Get error during save new order.", http.StatusInternalServerError)
		return
	}

	if isExist {
		// Is Existed for this user.
		isExist, err = u.market.IsExistForUser(req.Context(), userID, onumber)
		if err != nil {
			http.Error(res, "Get error during search duplicated order for user.", http.StatusInternalServerError)
			return
		}
		if isExist {
			// 202
			http.Error(res, "Order duplicated for User.", http.StatusAccepted)
			return
		}

		// Is Existed for others user.
		isExist, err = u.market.IsExistForOtherUsers(req.Context(), userID, onumber)
		if err != nil {
			http.Error(res, "Get error during search duplicated order for others.", http.StatusInternalServerError)
			return
		}

		if isExist {
			// 409
			http.Error(res, "Order duplicated for Other User.", http.StatusConflict)
			return
		}

	}

	// Add order to Observer for status upodating
	u.observer.AddOreder(order)

	// New number get to work 202
	res.WriteHeader(http.StatusAccepted)

	res.Write([]byte("Set order!"))

}

func (u *HandlerOrder) GetOrders(res http.ResponseWriter, req *http.Request) {

	//get UserID from cxt values
	ctxConfig := req.Context().Value(config.CtxConfig{}).(config.CtxConfig)

	// Check from middleware is user authorized
	if !ctxConfig.IsRegistered() {
		http.Error(res, "JWT not found.", http.StatusUnauthorized)
	}

	userID := ctxConfig.GetUserID()

	zap.S().Infoln("User id from middleware: ", userID)

	//set status code 200
	res.WriteHeader(http.StatusOK)

	res.Write([]byte("List of orders..."))

}
