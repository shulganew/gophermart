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

type AccrualResponce struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type Fetcher struct {
	stor FetcherUpdater
	conf *config.Config

	// map[order number]Order
	orders map[string]model.Order
	mu     sync.Mutex
}

type FetcherUpdater interface {
	LoadPocessing(ctx context.Context) ([]model.Order, error)
	UpdateStatus(ctx context.Context, order string, status model.Status) (err error)
	SetAccrual(ctx context.Context, order string, accrual decimal.Decimal) (err error)
	AddBonuses(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) (err error)
}

func NewFetcher(stor FetcherUpdater, conf *config.Config) *Fetcher {
	return &Fetcher{stor: stor, conf: conf, orders: make(map[string]model.Order, 0)}
}

func (o *Fetcher) AddOreder(order *model.Order) {
	o.mu.Lock()
	o.orders[order.OrderNr] = *order
	o.mu.Unlock()
}

func (o *Fetcher) Fetch(ctx context.Context) {
	upload := time.NewTicker(config.CheckAccrual * time.Second)
	go func(ctx context.Context, o *Fetcher) {
		for {
			<-upload.C
			o.FetchAccrual(ctx)
		}
	}(ctx, o)

}

// Load order data from database
func (o *Fetcher) FetchAccrual(ctx context.Context) {
	loadOrders, err := o.stor.LoadPocessing(ctx)
	if err != nil {
		zap.S().Errorln("Not all data was loaded to Fetcher... ", err)
	}

	for _, order := range loadOrders {
		// Set order status to PROCESSING in database
		o.stor.UpdateStatus(ctx, order.OrderNr, model.Status(model.PROCESSING))

		//fech status and accrual from Accrual system
		status, accrual, err := o.fetchOrderStatus(order.OrderNr)
		if err != nil {
			zap.S().Errorln("Get order status prepare error: ", err)
			continue
		}

		zap.S().Infoln("Get answer from Accrual system: ", "Order ", order, " status: ", status, " Accural: ", accrual)

		//if status PROCESSED or INVALID - update db and remove from orders
		if status == model.PROCESSED || status == model.INVALID {
			err = o.stor.UpdateStatus(ctx, order.OrderNr, status)
			if err != nil {
				zap.S().Errorln("Get error during deleted poccessed order", err)
			}
			//set accruals to the order
			if accrual != decimal.Zero {
				err = o.stor.SetAccrual(ctx, order.OrderNr, accrual)
				if err != nil {
					zap.S().Errorln("Get error during deleted poccessed order", err)
				}
			}

			//add accruals to user's bonus balance
			err = o.stor.AddBonuses(ctx, order.UserID, accrual)
			if err != nil {
				zap.S().Errorln("get error update user's balance", err)
			}

		}
	}
}

// Get data from Accrual system
func (o *Fetcher) fetchOrderStatus(orderNr string) (status model.Status, acc decimal.Decimal, err error) {

	client := &http.Client{}

	url, err := url.JoinPath(o.conf.Accrual, "api", "orders", orderNr)
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
