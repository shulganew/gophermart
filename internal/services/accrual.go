package services

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/entities"
	"go.uber.org/zap"
)

type AccrualService struct {
	stor          AccrualRepo
	accrualClient AccrualClient
}

type AccrualRepo interface {
	LoadPocessing(ctx context.Context) ([]entities.Order, error)
	UpdateStatus(ctx context.Context, order string, status entities.Status) (err error)
	SetAccrual(ctx context.Context, order string, accrual decimal.Decimal) (err error)
	AddBonuses(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) (err error)
}

type AccrualClient interface {
	GetOrderStatus(orderNr string) (*entities.AccrualResponce, error)
}

func NewAccrualService(accRepo AccrualRepo, ac AccrualClient) *AccrualService {
	return &AccrualService{stor: accRepo, accrualClient: ac}
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
		err := o.stor.UpdateStatus(ctx, order.OrderNr, entities.Status(entities.PROCESSING))
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

		status := entities.Status(accResp.Status)
		accrual := decimal.NewFromFloat(accResp.Accrual)

		zap.S().Infoln("Get answer from Accrual system: ", "Order ", order, " status: ", status, " Accural: ", accrual)

		//if status PROCESSED or INVALID - update db and remove from orders
		if status == entities.PROCESSED || status == entities.INVALID {
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
