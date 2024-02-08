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
	"github.com/shopspring/decimal"
	"github.com/shulganew/gophermart/internal/app"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/entities"
	"github.com/shulganew/gophermart/internal/services"
	"github.com/shulganew/gophermart/internal/services/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithdraw(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		requestURL string

		// all bonuses
		bonuses decimal.Decimal

		// all withdrawals
		withdrals decimal.Decimal

		// amount of withdrawn
		amount         decimal.Decimal
		Order          string
		statusCode     int
		orderIsExisted bool
	}{
		{
			name:           "Create withdrawn - order number (422), luna check",
			method:         http.MethodPost,
			Order:          "0265410804",
			requestURL:     "http://localhost:8080/api/user/balance/withdraw",
			bonuses:        decimal.NewFromFloat(12.2),
			withdrals:      decimal.NewFromFloat(6.2),
			amount:         decimal.NewFromFloat(1.0),
			statusCode:     http.StatusUnprocessableEntity,
			orderIsExisted: false,
		},
		{
			name:           "Create withdrawn - order number (422), Not found in database",
			method:         http.MethodPost,
			Order:          "7020147356",
			requestURL:     "http://localhost:8080/api/user/balance/withdraw",
			bonuses:        decimal.NewFromFloat(12.2),
			withdrals:      decimal.NewFromFloat(6.2),
			amount:         decimal.NewFromFloat(1.0),
			statusCode:     http.StatusUnprocessableEntity,
			orderIsExisted: true,
		},

		{
			name:           "Create withdrawn - 402 Payment Required",
			method:         http.MethodPost,
			Order:          "7020147356",
			requestURL:     "http://localhost:8080/api/user/balance/withdraw",
			bonuses:        decimal.NewFromFloat(6.2),
			withdrals:      decimal.NewFromFloat(62.2),
			amount:         decimal.NewFromFloat(100.0),
			statusCode:     http.StatusPaymentRequired,
			orderIsExisted: false,
		},

		{
			name:           "Create withdrawn - withdrawn sucsess",
			method:         http.MethodPost,
			Order:          "7020147356",
			requestURL:     "http://localhost:8080/api/user/balance/withdraw",
			bonuses:        decimal.NewFromFloat(12.2),
			withdrals:      decimal.NewFromFloat(6.2),
			amount:         decimal.NewFromFloat(1.0),
			statusCode:     http.StatusOK,
			orderIsExisted: false,
		},
	}

	app.InitLog()
	ctx := context.Background()

	conf := &config.Config{}

	// Init application
	conf.Address = "localhost:8088"
	conf.Accrual = "localhost:8090"
	conf.PassJWT = "JWTsecret"

	// init storage

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log("Test name: ", tt.name)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// crete mock storege
			repoUser := mocks.NewMockUserRepo(ctrl)
			repoCalc := mocks.NewMockCalcRepo(ctrl)
			repoOrder := mocks.NewMockOrderRepo(ctrl)

			register := services.NewUserService(repoUser)
			calc := services.NewCalcService(repoCalc)
			orderServ := services.NewOrderService(repoOrder)

			uuid, err := uuid.NewV7()
			assert.NoError(t, err)
			user := entities.User{UUID: uuid, Login: "Test123", Password: "123456"}

			_ = repoUser.EXPECT().
				AddUser(gomock.Any(), gomock.Any(), gomock.Any()).
				AnyTimes().
				Return(&user.UUID, nil)

			_ = repoOrder.EXPECT().
				AddOrder(gomock.Any(), gomock.Any()).
				AnyTimes().
				Return(nil)

			_ = repoOrder.EXPECT().
				IsExist(gomock.Any(), gomock.Any()).
				AnyTimes().
				Return(tt.orderIsExisted, nil)

			_ = repoCalc.EXPECT().
				MakeWithdrawn(gomock.Any(), gomock.Any(), gomock.Any()).
				AnyTimes().
				Return(nil)

			_ = repoCalc.EXPECT().
				GetBonuses(gomock.Any(), gomock.Any()).
				AnyTimes().
				Return(tt.bonuses, nil)

			_ = repoCalc.EXPECT().
				GetWithdrawals(gomock.Any(), gomock.Any()).
				AnyTimes().
				Return(tt.withdrals, nil)

			userID, exist, err := register.CreateUser(ctx, user.Login, user.Password)
			assert.NoError(t, err)
			assert.False(t, exist)

			wd := entities.Withdraw{OrderNr: tt.Order, Withdrawn: tt.amount.InexactFloat64()}

			jsonWs, err := json.Marshal(wd)
			if err != nil {
				log.Fatalln(err)
			}

			body := strings.NewReader(string(jsonWs))
			assert.NoError(t, err)

			// add chi context
			rctx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodPost, tt.requestURL, body)

			// add User and isRegister true tu context
			ctxUser := context.WithValue(req.Context(), entities.MiddlwDTO{}, entities.NewMiddlwDTO(*userID, true))

			req = req.WithContext(context.WithValue(ctxUser, chi.RouteCtxKey, rctx))

			jwt, _ := services.BuildJWTString(user.UUID, conf.PassJWT)

			req.Header.Add("Authorization", jwt)
			req.Header.Add("Content-Type", "application/json")

			// create status recorder
			resRecord := httptest.NewRecorder()

			// Make request
			balanceHand := NewHandlerBalance(conf, calc, orderServ)
			balanceHand.SetWithdraw(resRecord, req)

			// get result
			res := resRecord.Result()

			b, _ := io.ReadAll(res.Body)

			t.Log(string(b))

			err = res.Body.Close()
			assert.NoError(t, err)

			// check answer code
			t.Log("StatusCode test: ", tt.statusCode, " server: ", res.StatusCode)
			assert.Equal(t, tt.statusCode, res.StatusCode)
		})
	}
}

func TestBalance(t *testing.T) {
	tests := []struct {
		name       string
		requestURL string
		bonuses    decimal.Decimal
		withdrawn  decimal.Decimal
		statusCode int
	}{
		{
			name:       "Get Balans",
			requestURL: "http://localhost:8080/api/user/balance",
			bonuses:    decimal.NewFromFloat(12.2),
			withdrawn:  decimal.NewFromFloat(6.2),
			statusCode: http.StatusOK,
		},

		{
			name:       "Get balance 2",
			requestURL: "http://localhost:8080/api/user/balance",
			bonuses:    decimal.NewFromFloat(33.2),
			withdrawn:  decimal.NewFromFloat(22.2),
			statusCode: http.StatusOK,
		},
	}

	ctx := context.Background()

	conf := &config.Config{}

	// Init application
	conf.Address = "localhost:8088"
	conf.Accrual = "localhost:8090"
	conf.PassJWT = "JWTsecret"

	// init storage

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log("Test name: ", tt.name)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// crete mock storege
			repoUser := mocks.NewMockUserRepo(ctrl)
			repoCalc := mocks.NewMockCalcRepo(ctrl)
			repoOrder := mocks.NewMockOrderRepo(ctrl)

			userSrv := services.NewUserService(repoUser)
			calcSrv := services.NewCalcService(repoCalc)
			orderSrv := services.NewOrderService(repoOrder)

			uuid, err := uuid.NewV7()
			assert.NoError(t, err)
			user := entities.User{UUID: uuid, Login: "Test123", Password: "123"}

			_ = repoUser.EXPECT().
				AddUser(gomock.Any(), gomock.Any(), gomock.Any()).
				Times(1).
				Return(&user.UUID, nil)

			_ = repoCalc.EXPECT().
				GetBonuses(gomock.Any(), gomock.Any()).
				Times(1).
				Return(tt.bonuses, nil)

			_ = repoCalc.EXPECT().
				GetWithdrawn(gomock.Any(), gomock.Any()).
				Times(1).
				Return(tt.withdrawn, nil)

			userID, exist, err := userSrv.CreateUser(ctx, user.Login, user.Password)
			assert.NoError(t, err)
			assert.False(t, exist)

			// add chi context
			rctx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodGet, tt.requestURL, nil)

			// add User and isRegister true tu context
			ctxUser := context.WithValue(req.Context(), entities.MiddlwDTO{}, entities.NewMiddlwDTO(*userID, true))

			req = req.WithContext(context.WithValue(ctxUser, chi.RouteCtxKey, rctx))

			jwt, _ := services.BuildJWTString(user.UUID, conf.PassJWT)

			req.Header.Add("Authorization", jwt)
			req.Header.Add("Content-Type", "text/plain")

			// create status recorder
			resRecord := httptest.NewRecorder()

			// Make request
			balanceHand := NewHandlerBalance(conf, calcSrv, orderSrv)
			balanceHand.GetBalance(resRecord, req)

			// get result
			res := resRecord.Result()

			err = res.Body.Close()
			assert.NoError(t, err)

			// check answer code
			t.Log("StatusCode test: ", tt.statusCode, " server: ", res.StatusCode)
			assert.Equal(t, tt.statusCode, res.StatusCode)

			// Unmarshal body

			var balance entities.UserBalance
			err = json.NewDecoder(res.Body).Decode(&balance)
			require.NoError(t, err)

			b := decimal.NewFromFloat(balance.Bonus)
			w := decimal.NewFromFloat(balance.Withdrawn)

			bt := tt.bonuses
			wt := tt.withdrawn

			assert.Equal(t, b.Equal(bt), true)
			assert.Equal(t, w.Equal(wt), true)

			t.Log("StatusCode test: ", tt.statusCode, " server: ", res.StatusCode)
			assert.Equal(t, tt.statusCode, res.StatusCode)
		})
	}
}
