package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/model"
	"github.com/shulganew/gophermart/internal/services"
	"github.com/shulganew/gophermart/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestUserLogin(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		requestURL string
		login      string
		passLogin  string
		passDB     string
		statusCode int
	}{
		{
			name:       "Login user success",
			method:     http.MethodPost,
			requestURL: "http://localhost:8080/api/user/login",
			login:      "user",
			passLogin:  "qwerty",
			passDB:     "qwerty",
			statusCode: http.StatusOK,
		},
		{
			name:       "Login user success",
			method:     http.MethodPost,
			requestURL: "http://localhost:8080/api/user/login",
			login:      "user",
			passLogin:  "qwerty",
			passDB:     "asdfg",
			statusCode: http.StatusUnauthorized,
		},
	}

	conf := &config.Config{}

	//Init application
	//market, register, observer := app.InitApp(ctx, conf, db)
	conf.Address = "localhost:8088"
	conf.Accrual = "localhost:8090"
	conf.PassJWT = "JWTsecret"

	//init storage

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			t.Log("Test name: ", tt.name)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			//crete mock storege
			repoRegister := mocks.NewMockRegistrar(ctrl)

			register := services.NewRegister(repoRegister)

			uuid, err := uuid.NewV7()
			assert.NoError(t, err)

			cPass, err := bcrypt.GenerateFromPassword([]byte(tt.passDB), bcrypt.DefaultCost)
			assert.NoError(t, err)

			dbUser := model.User{UUID: &uuid, Login: tt.login, Password: string(cPass)}

			_ = repoRegister.EXPECT().
				GetByLogin(gomock.Any(), gomock.Any()).
				AnyTimes().
				Return(&dbUser, nil)

			loginUser := model.User{UUID: &uuid, Login: tt.login, Password: tt.passLogin}

			jsonWs, err := json.Marshal(loginUser)
			if err != nil {
				log.Fatalln(err)
			}

			body := strings.NewReader(string(jsonWs))
			assert.NoError(t, err)

			//add chi context
			rctx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodPost, tt.requestURL, body)

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			req.Header.Add("Content-Type", "application/json")

			//create status recorder
			resRecord := httptest.NewRecorder()

			//Make request
			userLogin := NewHandlerLogin(conf, register)
			userLogin.LoginUser(resRecord, req)

			//get result
			res := resRecord.Result()

			b, _ := io.ReadAll(res.Body)

			t.Log(string(b))

			defer res.Body.Close()

			//check answer code
			t.Log("StatusCode test: ", tt.statusCode, " server: ", res.StatusCode)

			//Unmarshal body

			assert.Equal(t, tt.statusCode, res.StatusCode)

		})
	}
}
