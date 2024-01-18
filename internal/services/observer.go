package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/model"
	"go.uber.org/zap"
)

// Check Acceral service every X sec
const CheckAccural = 3

// Check Oraders in DB every X sec
const UploadData = 3

type AccrualResponce struct {
	Order   string `json:"order"`
	Status  string `json:"status"`
	Accrual string `json:"accrual"`
}

type Observer struct {
	stor ObserverUpdater
	conf *config.Config

	// map[order number]Order
	orders map[string]*model.Order
	mu     sync.Mutex
}

type ObserverUpdater interface {
	// SetOrder(ctx context.Context, userID *uuid.UUID, order string) error
	// IsExistForUser(ctx context.Context, userID *uuid.UUID, order string) (isExist bool, err error)
	// IsExistForOtherUsers(ctx context.Context, userID *uuid.UUID, order string) (isExist bool, err error)
	LoadOrders(ctx context.Context) ([]model.Order, error)
	UpdateOrderStatus(ctx context.Context, order *model.Order, status *model.Status, accural *decimal.Decimal) error
}

func NewObserver(stor ObserverUpdater, conf *config.Config) *Observer {
	return &Observer{stor: stor, conf: conf, orders: make(map[string]*model.Order, 0)}
}

func (o *Observer) AddOreder(order *model.Order) {
	o.mu.Lock()
	o.orders[order.Onumber] = order
	o.mu.Unlock()
}

func (o *Observer) dellOreder(order *model.Order) {
	o.mu.Lock()
	//TODO
	o.mu.Unlock()
}

func (o *Observer) Observ(ctx context.Context) {

	observ := time.NewTicker(CheckAccural * time.Second)
	upload := time.NewTicker(UploadData * time.Second)
	go func(ctx context.Context, o *Observer) {
		for {
			select {
			case <-observ.C:
				o.ObservAccural(ctx)
			case <-upload.C:
				o.LoadData(ctx)
			}
		}
	}(ctx, o)

}

func (o *Observer) ObservAccural(ctx context.Context) {

	o.mu.Lock()
	zap.S().Infoln("ObservAccural lenth order: ", len(o.orders))
	for _, order := range o.orders {

		status, accural, err := o.getOrderStatus(order)
		if err != nil {
			zap.S().Errorln("Get order status prepare error ", err)
		}
		//if status PROCESSED or INVALID - update db and remove from orders
		if *status == 2 || *status == 3 {
			err := o.stor.UpdateOrderStatus(ctx, order, status, accural)
			if err != nil {
				zap.S().Errorln("Get error during update", err)
			}
			//remove order from memory map
			delete(o.orders, order.Onumber)
		}
	}
	o.mu.Unlock()

}

// Load order data from database
func (o *Observer) LoadData(ctx context.Context) {
	loadOrders, err := o.stor.LoadOrders(ctx)
	if err != nil {
		zap.S().Errorln("Not all data was loaded... ", err)
	}

	o.mu.Lock()
	for _, order := range loadOrders {
		// Add order if it not existe in Observer
		if _, ok := o.orders[order.Onumber]; !ok {
			zap.S().Infoln("Load order from database: ", order.Onumber)
			o.orders[order.Onumber] = &order
		}
	}
	o.mu.Unlock()
}

// Get data from Accural system
func (o *Observer) getOrderStatus(order *model.Order) (status *model.Status, accural *decimal.Decimal, err error) {

	client := &http.Client{}

	url, err := url.JoinPath("http://", o.conf.Accrual, "api", "orders", order.Onumber)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	if res.StatusCode != http.StatusOK {
		return &order.Status, nil, nil
	}

	//Load data to AccrualResponce from json
	var accResp AccrualResponce
	err = json.NewDecoder(res.Body).Decode(&accResp)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()

	st := model.Status(0)
	st.SetStatus(accResp.Status)
	var accrual decimal.Decimal
	if accResp.Accrual != "" {
		accrual, err = decimal.NewFromString(accResp.Accrual)
		if err != nil {
			zap.S().Errorln("Error create decimal from string ", accResp.Accrual, err)
		}
	}

	return &st, &accrual, nil
}
