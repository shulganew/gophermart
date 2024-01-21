package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/model"
	"github.com/shulganew/gophermart/internal/service/mocks"
	"github.com/shulganew/gophermart/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBalance(t *testing.T) {
	tests := []struct {
		name       string
		requestURL string
		acc        *decimal.Decimal
		withdrawn  *decimal.Decimal
		statusCode int
		//want
	}{
		{
			name:       "Get Banans",
			requestURL: "http://localhost:8080/api/user/balance",
			acc:        &[]decimal.Decimal{decimal.NewFromFloat(12.2)}[0],
			withdrawn:  &[]decimal.Decimal{decimal.NewFromFloat(6.2)}[0],
			statusCode: http.StatusOK,
		},

		{
			name:       "Get newgative balance",
			requestURL: "http://localhost:8080/api/user/balance",
			acc:        &[]decimal.Decimal{decimal.NewFromFloat(33.2)}[0],
			withdrawn:  &[]decimal.Decimal{decimal.NewFromFloat(22.2)}[0],
			statusCode: http.StatusOK,
		},
	}

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

			t.Log("=============Test Balance===============")

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			//crete mock storege

			repoRegister := mocks.NewMockRegistrar(ctrl)
			repoMarket := mocks.NewMockMarketPlaceholder(ctrl)

			register := services.NewRegister(repoRegister)
			market := services.NewMarket(repoMarket)

			uuid, err := uuid.NewV7()
			assert.NoError(t, err)
			user := model.User{UUID: &uuid, Login: "Test123", Password: "123"}

			_ = repoRegister.EXPECT().
				Register(gomock.Any(), gomock.Any()).
				Times(1).
				Return(nil)

			_ = repoMarket.EXPECT().
				GetAccruals(gomock.Any(), gomock.Any()).
				Times(1).
				Return(tt.acc, nil)

			_ = repoMarket.EXPECT().
				GetWithdrawns(gomock.Any(), gomock.Any()).
				Times(1).
				Return(tt.withdrawn, nil)

			userID, exist, err := register.NewUser(ctx, user)
			assert.NoError(t, err)
			assert.False(t, exist)

			//add chi context
			rctx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodGet, tt.requestURL, nil)

			// add User and isRegister true tu context
			ctxUser := context.WithValue(req.Context(), config.CtxConfig{}, config.NewCtxConfig(userID, true))

			req = req.WithContext(context.WithValue(ctxUser, chi.RouteCtxKey, rctx))

			jwt, _ := services.BuildJWTString(&user, conf.PassJWT)

			req.Header.Add("Authorization", jwt)
			req.Header.Add("Content-Type", "text/plain")

			//create status recorder
			resRecord := httptest.NewRecorder()

			//Make request
			balanceHand := NewHandlerBalance(conf, market)
			balanceHand.GetBalance(resRecord, req)

			//get result
			res := resRecord.Result()
			defer res.Body.Close()

			//check answer code
			t.Log("StatusCode test: ", tt.statusCode, " server: ", res.StatusCode)
			assert.Equal(t, tt.statusCode, res.StatusCode)

			//Unmarshal body

			var balance model.UserBalance
			err = json.NewDecoder(res.Body).Decode(&balance)
			require.NoError(t, err)

			b := decimal.NewFromFloat(balance.Bonus)
			w := decimal.NewFromFloat(balance.Withdrawn)

			bt := *tt.acc
			wt := *tt.withdrawn

			assert.Equal(t, b.Equal(bt.Sub(w)), true)
			assert.Equal(t, w.Equal(wt), true)

			t.Log("StatusCode test: ", tt.statusCode, " server: ", res.StatusCode)
			assert.Equal(t, tt.statusCode, res.StatusCode)

		})
	}
}
