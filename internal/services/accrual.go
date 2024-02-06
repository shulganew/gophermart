package services

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/shulganew/gophermart/internal/accrual"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/model"
	"go.uber.org/zap"
)

type AccrualService struct {
	stor          AccrualRepo
	conf          *config.Config
	accrualClient *accrual.AccrualClient
}

type AccrualRepo interface {
	LoadPocessing(ctx context.Context) ([]model.Order, error)
	UpdateStatus(ctx context.Context, order string, status model.Status) (err error)
	SetAccrual(ctx context.Context, order string, accrual decimal.Decimal) (err error)
	AddBonuses(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) (err error)
}

func NewAccrualService(accRepo AccrualRepo, conf *config.Config, ac *accrual.AccrualClient) *AccrualService {
	return &AccrualService{stor: accRepo, conf: conf, accrualClient: ac}
}

func (o *AccrualService) Run(ctx context.Context) {
	upload := time.NewTicker(config.CheckAccrual * time.Second)
	go func(ctx context.Context, o *AccrualService) {
		for {
			<-upload.C
			o.FetchAccrual(ctx)
		}
	}(ctx, o)
}

func (o *AccrualService) FetchAccrual(ctx context.Context) {
	loadOrders, err := o.stor.LoadPocessing(ctx)
	if err != nil {
		zap.S().Errorln("Not all data was loaded to Fetcher... ", err)
	}

	for _, order := range loadOrders {
		// Set order status to PROCESSING in database
		err := o.stor.UpdateStatus(ctx, order.OrderNr, model.Status(model.PROCESSING))
		if err != nil {
			zap.S().Errorln("Can't update status to PROCESSING in database", err)
			continue
		}
		//fech status and accrual from Accrual system
		accResp, err := o.accrualClient.GetOrderStatus(order.OrderNr)
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
