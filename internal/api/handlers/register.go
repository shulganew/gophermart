package handlers

import (
	"encoding/json"

	"net/http"

	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/entities"
	"github.com/shulganew/gophermart/internal/services"
	"go.uber.org/zap"
)

type HandlerRegister struct {
	userSrv *services.UserService
	conf    *config.Config
}

func NewHandlerRegister(conf *config.Config, usrSrv *services.UserService) *HandlerRegister {
	return &HandlerRegister{userSrv: usrSrv, conf: conf}
}

// Adding new user to Market.
func (u *HandlerRegister) SetUser(res http.ResponseWriter, req *http.Request) {
	var user entities.User

	if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
		// If can't decode 400
		errt := "Can't decode JSON"
		zap.S().Errorln(errt, err)
		http.Error(res, errt, http.StatusBadRequest)
		return
	}
	zap.S().Infoln("New user:", user.Login)

	userID, exist, err := u.userSrv.CreateUser(req.Context(), user.Login, user.Password)
	if err != nil {
		// If can't get UUID or hash pass 500
		errt := "Can't get UUID or user hash"
		zap.S().Errorln(errt, err)
		http.Error(res, errt, http.StatusInternalServerError)
		return
	}
	if exist {
		// 409 - login is used
		errt := "Users login is used"
		zap.S().Infoln(errt, err)
		http.Error(res, errt, http.StatusConflict)
		return
	}

	user.UUID = *userID

	jwt, _ := services.BuildJWTString(user.UUID, u.conf.PassJWT)

	res.Header().Add("Content-Type", "text/plain")

	res.Header().Add("Authorization", jwt)

	// set status code 200
	res.WriteHeader(http.StatusOK)

	_, err = res.Write([]byte("User added."))
	if err != nil {
		zap.S().Errorln("Can't write to response in SetUser handler", err)
	}
}
