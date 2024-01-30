package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/shulganew/gophermart/internal/app"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/model"
	"github.com/shulganew/gophermart/internal/services"
	"github.com/shulganew/gophermart/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrders(t *testing.T) {
	tests := []struct {
		name       string
		requestURL string
		orders     []model.Order
		statusCode int
		//want
	}{
		{
			name:       "Get 2 Orders",
			requestURL: "http://localhost:8080/api/user/orders",

			orders: getOrders(),

			statusCode: http.StatusOK,
		},

		{
			name: "No Orders",

			requestURL: "http://localhost:8080/api/user/orders",
			orders:     make([]model.Order, 0),
			statusCode: http.StatusNoContent,
		},
	}

	app.InitLog()
	ctx := context.Background()

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
			repoMarket := mocks.NewMockMarketPlaceholder(ctrl)
			repoObserver := mocks.NewMockObserverUpdater(ctrl)

			register := services.NewRegister(repoRegister)
			market := services.NewMarket(repoMarket)
			observer := services.NewObserver(repoObserver, conf)

			uuid, err := uuid.NewV7()
			assert.NoError(t, err)
			user := model.User{UUID: &uuid, Login: "Test123", Password: "123"}

			_ = repoRegister.EXPECT().
				AddUser(gomock.Any(), gomock.Any(), gomock.Any()).
				Times(1).
				Return(user.UUID, nil)

			_ = repoMarket.EXPECT().
				GetOrders(gomock.Any(), gomock.Any()).
				Times(1).
				Return(tt.orders, nil)

			userID, exist, err := register.NewUser(ctx, user.Login, user.Password)
			assert.NoError(t, err)
			assert.False(t, exist)

			//add chi context
			rctx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodGet, tt.requestURL, nil)

			// add User and isRegister true tu context
			ctxUser := context.WithValue(req.Context(), config.CtxConfig{}, config.NewCtxConfig(userID, true))

			req = req.WithContext(context.WithValue(ctxUser, chi.RouteCtxKey, rctx))

			jwt, _ := services.BuildJWTString(userID, conf.PassJWT)

			req.Header.Add("Authorization", jwt)
			req.Header.Add("Content-Type", "text/plain")

			//create status recorder
			resRecord := httptest.NewRecorder()

			ordersHand := NewHandlerOrder(conf, market, observer)
			ordersHand.GetOrders(resRecord, req)

			//get result
			res := resRecord.Result()
			defer res.Body.Close()

			//check answer code
			t.Log("StatusCode test: ", tt.statusCode, " server: ", res.StatusCode)
			assert.Equal(t, tt.statusCode, res.StatusCode)

			//Unmarshal body

			if res.StatusCode != http.StatusNoContent {
				var responses []OrderResponse
				err = json.NewDecoder(res.Body).Decode(&responses)
				require.NoError(t, err)

				for _, response := range responses {
					t.Log(response.Onumber)
					t.Log(response.Status)
					t.Log(response.Accrual)
					t.Log(response.Uploaded)
				}
			}

			t.Log("StatusCode test: ", tt.statusCode, " server: ", res.StatusCode)
			assert.Equal(t, tt.statusCode, res.StatusCode)

		})
	}
}

func getOrders() []model.Order {
	userID, err := uuid.DefaultGenerator.NewV7()
	if err != nil {
		log.Fatalln(err)
	}
	return []model.Order{*model.NewOrder(&userID, goluhn.Generate(10), false, decimal.NewFromFloat(20), decimal.NewFromFloat(200)),
		*model.NewOrder(&userID, goluhn.Generate(10), false, decimal.NewFromFloat(5), decimal.NewFromFloat(100))}
}
