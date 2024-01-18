package services

import (
	"context"
	"errors"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shulganew/gophermart/internal/model"
	"go.uber.org/zap"
)

// User registration service
type Market struct {
	stor MarketPlaceholder
}

type MarketPlaceholder interface {
	SetOrder(ctx context.Context, userID *uuid.UUID, order string) error
	IsExistForUser(ctx context.Context, userID *uuid.UUID, order string) (isExist bool, err error)
	IsExistForOtherUsers(ctx context.Context, userID *uuid.UUID, order string) (isExist bool, err error)
}

func NewMarket(stor MarketPlaceholder) *Market {
	return &Market{stor: stor}
}

func (m *Market) SetOrder(ctx context.Context, order *model.Order) (existed bool, err error) {

	// Add order to the database.
	err = m.stor.SetOrder(ctx, order.UserID, order.Onumber)
	if err != nil {
		var pgErr *pgconn.PgError
		// If Order exist in the DataBase
		if errors.As(err, &pgErr) && pgerrcode.UniqueViolation == pgErr.Code {
			zap.S().Infoln("Order exist: ", order)
			return true, err
		}
		zap.S().Errorln("Set order error: ", order)
	}

	return false, nil
}

func (m *Market) IsExistForUser(ctx context.Context, userID *uuid.UUID, order string) (isExist bool, err error) {
	return m.stor.IsExistForUser(ctx, userID, order)
}

func (m *Market) IsExistForOtherUsers(ctx context.Context, userID *uuid.UUID, order string) (isExist bool, err error) {
	return m.stor.IsExistForOtherUsers(ctx, userID, order)
}
