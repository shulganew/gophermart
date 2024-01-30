package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/model"
	"go.uber.org/zap"
)

// Check Acceral service every X sec
const CheckAccrual = 1

// Check Oraders in DB every X sec
const UploadData = 3

type AccrualResponce struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type Observer struct {
	stor ObserverUpdater
	conf *config.Config

	// map[order number]Order
	orders map[string]model.Order
	mu     sync.Mutex
}

type ObserverUpdater interface {
	LoadPocessing(ctx context.Context) ([]model.Order, error)
	UpdateStatus(ctx context.Context, order string, status model.Status) (err error)
	SetAccrual(ctx context.Context, order string, accrual decimal.Decimal) (err error)
	AddBonuses(ctx context.Context, userID *uuid.UUID, amount decimal.Decimal) (err error)
}

func NewObserver(stor ObserverUpdater, conf *config.Config) *Observer {
	return &Observer{stor: stor, conf: conf, orders: make(map[string]model.Order, 0)}
}

func (o *Observer) AddOreder(order *model.Order) {
	o.mu.Lock()
	o.orders[order.Onumber] = *order
	o.mu.Unlock()
}

func (o *Observer) Observ(ctx context.Context) {

	observ := time.NewTicker(CheckAccrual * time.Second)
	upload := time.NewTicker(UploadData * time.Second)
	go func(ctx context.Context, o *Observer) {
		for {
			select {
			case <-observ.C:
				o.ObservAccrual(ctx)
			case <-upload.C:
				o.LoadData(ctx)
			}
		}
	}(ctx, o)

}

func (o *Observer) ObservAccrual(ctx context.Context) {
	o.mu.Lock()

	for _, order := range o.orders {

		status, accrual, err := o.getOrderStatus(&order)
		if err != nil {
			zap.S().Errorln("Get order status prepare error: ", err)
			continue
		}

		zap.S().Infoln("Get answer from Accrual system: ", "Order ", order.Onumber, " status: ", status, " Accural: ", accrual)

		//if status PROCESSED or INVALID - update db and remove from orders
		if status == model.PROCESSED || status == model.INVALID {
			err = o.stor.UpdateStatus(ctx, order.Onumber, status)
			if err != nil {
				zap.S().Errorln("Get error during deleted poccessed order", err)
			}
			//set accruals to the order
			if accrual != decimal.Zero {
				err = o.stor.SetAccrual(ctx, order.Onumber, accrual)
				if err != nil {
					zap.S().Errorln("Get error during deleted poccessed order", err)
				}
			}

			//add accruals to user's bonus balance
			err = o.stor.AddBonuses(ctx, order.UserID, accrual)
			if err != nil {
				zap.S().Errorln("get error update user's balance", err)
			}
			delete(o.orders, order.Onumber)

		}
	}

	o.mu.Unlock()
}

// Load order data from database
func (o *Observer) LoadData(ctx context.Context) {
	loadOrders, err := o.stor.LoadPocessing(ctx)
	if err != nil {
		zap.S().Errorln("Not all data was loaded... ", err)
	}

	o.mu.Lock()
	for _, order := range loadOrders {

		// Add order if it not existe in Observer
		if _, ok := o.orders[order.Onumber]; !ok {
			zap.S().Infoln("Load order from database: ", order.Onumber)
			o.orders[order.Onumber] = order

			// Set order status to PROCESSING in database
			order.Status = model.Status(model.PROCESSING)
			o.stor.UpdateStatus(ctx, order.Onumber, order.Status)
		}
	}
	o.mu.Unlock()

}

// Get data from Accrual system
func (o *Observer) getOrderStatus(order *model.Order) (status model.Status, acc decimal.Decimal, err error) {

	client := &http.Client{}

	url, err := url.JoinPath(o.conf.Accrual, "api", "orders", order.Onumber)
	if err != nil {
		return "", decimal.Zero, err
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", decimal.Zero, err
	}

	res, err := client.Do(req)
	if err != nil {
		return "", decimal.Zero, err
	}

	// Set status INVALID if no content 204
	if res.StatusCode == http.StatusNoContent {
		invalid := model.Status(model.INVALID)
		return invalid, decimal.Zero, nil
	}

	// Set status PROCESSING if no busy 429
	if res.StatusCode == http.StatusNoContent {
		processing := model.Status(model.PROCESSING)
		return processing, decimal.Zero, nil
	}

	//Load data to AccrualResponce from json
	var accResp AccrualResponce
	err = json.NewDecoder(res.Body).Decode(&accResp)
	if err != nil {
		return "", decimal.Zero, err
	}
	defer res.Body.Close()
	st := model.Status(accResp.Status)
	accrual := decimal.NewFromFloat(accResp.Accrual)

	return st, accrual, nil
}
