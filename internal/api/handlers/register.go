package handlers

import (
	"encoding/json"

	"net/http"

	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/model"
	"github.com/shulganew/gophermart/internal/services"
	"go.uber.org/zap"
)

type HandlerRegister struct {
	register *services.Register
	conf     *config.Config
}

func NewHandlerRegister(conf *config.Config, register *services.Register) *HandlerRegister {

	return &HandlerRegister{register: register, conf: conf}
}

// Adding new user to Market
func (u *HandlerRegister) SetUser(res http.ResponseWriter, req *http.Request) {

	var user model.User

	if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
		// If can't decode 400
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	zap.S().Infoln("New user:", user.Login)

	userID, exist, err := u.register.NewUser(req.Context(), user)
	if err != nil {
		// If can't get UUID or hash pass 500
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
	if exist {
		// 409 - login is used
		http.Error(res, "Uesr existed", http.StatusConflict)
		return
	}

	user.UUID = userID

	jwt, _ := services.BuildJWTString(&user, u.conf.PassJWT)

	res.Header().Add("Content-Type", "text/plain")

	res.Header().Add("Authorization", jwt)

	//set status code 200
	res.WriteHeader(http.StatusOK)

	res.Write([]byte("User added."))

}
