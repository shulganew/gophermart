package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/model"
	"github.com/shulganew/gophermart/internal/services"
	"go.uber.org/zap"
)

type HandlerLogin struct {
	register *services.Register
	conf     *config.Config
}

func NewHandlerLogin(conf *config.Config, register *services.Register) *HandlerLogin {

	return &HandlerLogin{register: register, conf: conf}
}

// Adding new user to Market
func (u *HandlerLogin) LoginUser(res http.ResponseWriter, req *http.Request) {

	var user model.User

	if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
		// If can't decode 400
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	zap.S().Infoln("Login user:", user.Login)
	zap.S().Infoln("Login Password:", user.Password)

	userID, isValid := u.register.IsValid(req.Context(), user)
	if !isValid {
		// Wrond user login or password 401
		http.Error(res, "Wrong login or password", http.StatusUnauthorized)
		return
	}

	zap.S().Infoln("Login sucsess, user id is: ", userID)

	jwt, _ := u.register.BuildJWTString(user)
	zap.S().Infoln("Create JWT: ", jwt)

	res.Header().Add("Content-Type", "text/plain")

	res.Header().Add("Authorization", "Bearer "+jwt)

	//set status code 200
	res.WriteHeader(http.StatusOK)

	res.Write([]byte("User loged in."))

}
