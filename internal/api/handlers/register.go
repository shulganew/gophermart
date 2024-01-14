package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/model"
	"github.com/shulganew/gophermart/internal/services"
)

type HandlerRegister struct {
	register *services.Register
	conf     *config.Config
}

func NewHandlerRegister(conf *config.Config, register *services.Register) *HandlerRegister {

	return &HandlerRegister{register: register, conf: conf}
}

// POTS and new User to Market
func (u *HandlerRegister) SetUser(res http.ResponseWriter, req *http.Request) {

	var user model.User

	if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Header().Add("Content-Type", "text/plain")

	//set status code 200
	res.WriteHeader(http.StatusOK)

	res.Write([]byte("User added."))

}
