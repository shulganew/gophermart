package services

import (
	"context"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/shulganew/gophermart/internal/api/client"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/model"
	"go.uber.org/zap"
)

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
		accResp, err := client.FetchOrderStatus(order.OrderNr, o.conf)
		if err != nil {
			zap.S().Errorln("Get order status prepare error: ", err)
			continue
		}

		status := model.Status(accResp.Status)
		accrual := decimal.NewFromFloat(accResp.Accrual)

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
