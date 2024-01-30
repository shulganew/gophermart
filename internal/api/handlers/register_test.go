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
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"github.com/shulganew/gophermart/internal/app"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/model"
	"github.com/shulganew/gophermart/internal/services"
	"github.com/shulganew/gophermart/internal/services/mocks"
	"github.com/stretchr/testify/assert"
)

func TestUserRegister(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		requestURL    string
		login         string
		passLogin     string
		statusCode    int
		registerError error
	}{
		{
			name:          "Login user success",
			method:        http.MethodPost,
			requestURL:    "http://localhost:8080/api/user/register",
			login:         "user",
			passLogin:     "qwerty123456",
			statusCode:    http.StatusOK,
			registerError: nil,
		},
		{
			name:          "Registration duplicated user",
			method:        http.MethodPost,
			requestURL:    "http://localhost:8080/api/user/register",
			login:         "user",
			passLogin:     "qwerty123456",
			statusCode:    http.StatusConflict,
			registerError: &pq.Error{Code: pq.ErrorCode(pgerrcode.UniqueViolation)},
		},
	}

	app.InitLog()
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

			user := model.User{UUID: &uuid, Login: tt.login, Password: string(tt.passLogin)}

			_ = repoRegister.EXPECT().
				AddUser(gomock.Any(), gomock.Any(), gomock.Any()).
				AnyTimes().
				Return(tt.registerError)

			_ = repoRegister.EXPECT().
				GetByLogin(gomock.Any(), gomock.Any()).
				AnyTimes().
				Return(&user, tt.registerError)

			jsonWs, err := json.Marshal(user)
			if err != nil {
				log.Fatalln(err)
			}

			body := strings.NewReader(string(jsonWs))
			assert.NoError(t, err)

			//add chi context
			rctx := chi.NewRouteContext()
			req := httptest.NewRequest(tt.method, tt.requestURL, body)

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			req.Header.Add("Content-Type", "application/json")

			//create status recorder
			resRecord := httptest.NewRecorder()

			//Make request
			regUser := NewHandlerRegister(conf, register)
			regUser.SetUser(resRecord, req)

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
