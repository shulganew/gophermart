package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/shulganew/gophermart/internal/model"
)

func (base *Repo) AddOrder(ctx context.Context, userID uuid.UUID, order string, isPreOrder bool, withdrawn decimal.Decimal) error {
	query := `
	INSERT INTO orders (user_id, order_number, is_preorder, uploaded, withdrawn) 
	VALUES ($1, $2, $3, $4, $5)
	`
	_, err := base.master.ExecContext(ctx, query, userID, order, isPreOrder, time.Now(), withdrawn)
	if err != nil {
		var pgErr *pq.Error
		// if order exist in DataBase
		if errors.As(err, &pgErr) && pgerrcode.UniqueViolation == pgErr.Code {
			return pgErr
		}
		return fmt.Errorf("error during set order to Storage, error: %w", err)
	}

	return nil
}

func (base *Repo) GetOrders(ctx context.Context, userID uuid.UUID) ([]model.Order, error) {
	query := `
	SELECT user_id, order_number, uploaded, status, withdrawn, accrual
	FROM orders 
	WHERE is_preorder = FALSE AND user_id = $1
	ORDER BY uploaded DESC
	`
	orders := []model.Order{}
	err := base.master.MustBegin().SelectContext(ctx, &orders, query, userID)
	if err != nil {
		return nil, err
	}
	return orders, nil

}

func (base *Repo) IsExist(ctx context.Context, order string) (isExist bool, err error) {
	query := `
	SELECT count(*) 
	FROM orders 
	WHERE order_number = $1
	`
	var ordersn int
	err = base.master.GetContext(ctx, &ordersn, query, order)
	if err != nil {
		return true, fmt.Errorf("error during order search for user: %w", err)
	}
	return ordersn != 0, nil
}

func (base *Repo) IsExistForOtherUser(ctx context.Context, userID uuid.UUID, order string) (isExist bool, err error) {
	query := `
	SELECT count(*) 
	FROM orders WHERE user_id != $1 
	AND order_number = $2
	`
	var ordersn int
	err = base.master.GetContext(ctx, &ordersn, query, userID, order)
	if err != nil {
		return true, fmt.Errorf("error during order search for user: %w", err)
	}
	return ordersn != 0, nil
}

func (base *Repo) GetWithdrawals(ctx context.Context, userID uuid.UUID) (withdrawn decimal.Decimal, err error) {
	query := `
	SELECT withdrawals 
	FROM users 
	WHERE user_id = $1
	`
	err = base.master.GetContext(ctx, &withdrawn, query, userID)
	if err != nil {
		return decimal.Zero, err
	}
	return
}

func (base *Repo) Withdrawals(ctx context.Context, userID uuid.UUID) (wds []model.Withdrawals, err error) {
	query := `
	SELECT  order_number, withdrawn, uploaded
		FROM orders 
		WHERE user_id = $1
		ORDER BY uploaded DESC
	`
	err = base.master.MustBegin().SelectContext(ctx, &wds, query, userID)
	if err != nil {
		return nil, err
	}
	return
}

func (base *Repo) IsPreOrder(ctx context.Context, userID uuid.UUID, order string) (bool, error) {
	query := `
	SELECT count(order_number)
	FROM orders
	WHERE user_id = $1 AND order_number = $2 AND is_preorder = TRUE
	`
	var is int
	err := base.master.GetContext(ctx, &is, query, userID, order)
	if err != nil {
		return true, fmt.Errorf("error during is preorder checking: %w", err)
	}

	return is != 0, err
}

// Move preorder to regular order. Add accruals for this order.
func (base *Repo) MovePreOrder(ctx context.Context, order *model.Order) (err error) {
	query := `
	UPDATE orders 
	SET status = $1, is_preorder = $2 
	WHERE order_number = $3
	`
	_, err = base.master.ExecContext(ctx, query, order.Status, order.IsPreOrder, order.OrderNr)
	if err != nil {
		return fmt.Errorf("move order error, can't move preoreder to order, %w", err)
	}

	return

}

func (base *Repo) GetBonuses(ctx context.Context, userID uuid.UUID) (accrual decimal.Decimal, err error) {
	query := `
	SELECT bonuses 
	FROM users 
	WHERE user_id = $1
	`
	err = base.master.GetContext(ctx, &accrual, query, userID)
	if err != nil {
		return decimal.Zero, err
	}
	return
}
func (base *Repo) GetWithdrawn(ctx context.Context, userID uuid.UUID) (wd decimal.Decimal, err error) {
	query := `
	SELECT withdrawals
	FROM users 
	WHERE user_id = $1
	`
	err = base.master.GetContext(ctx, &wd, query, userID)
	if err != nil {
		return decimal.Zero, err
	}
	return
}

func (base *Repo) SetAccrual(ctx context.Context, order string, accrual decimal.Decimal) (err error) {
	query := `
	UPDATE orders 
	SET accrual = $1 
	WHERE order_number = $2
	`
	_, err = base.master.ExecContext(ctx, query, accrual, order)
	if err != nil {

		return fmt.Errorf("can't update order's accrual during update status %w", err)
	}

	return
}

func (base *Repo) AddBonuses(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) (err error) {
	query := `
	UPDATE users 
	SET bonuses = bonuses + $1 
	WHERE user_id = $2
	`
	_, err = base.master.ExecContext(ctx, query, amount, userID)
	if err != nil {
		return fmt.Errorf("can't update add to user's order accruals to bonuses %w", err)
	}

	return
}

// Move user's amount from bonuses to withdrawals.
func (base *Repo) MakeWithdrawn(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) (err error) {
	queryBonusUpdate := `
	UPDATE users 
	SET bonuses = bonuses - $1 
	WHERE user_id = $2
	`
	tx := base.master.MustBegin()
	_, err = tx.ExecContext(ctx, queryBonusUpdate, amount, userID)
	if err != nil {
		return fmt.Errorf("can't make bonuse withdrawn, %w", err)
	}

	queryBonusCheck := `
	SELECT bonuses 
	FROM users 
	WHERE user_id = $1
	`
	//check uses balance after update
	var bonuses decimal.Decimal

	//check uses balance in transaction
	err = tx.GetContext(ctx, &bonuses, queryBonusCheck, userID)
	if err != nil || bonuses.IsNegative() {
		tx.Rollback()
		return fmt.Errorf("error during user's bonuse withdrawn: %w", err)
	}

	queryWithdrawnUpdate := `
	UPDATE users 
	SET withdrawals = withdrawals + $1 
	WHERE user_id = $2
	`
	//Update user's withdrawals
	_, err = tx.ExecContext(ctx, queryWithdrawnUpdate, amount, userID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("can't add withdrawns to user, %w", err)
	}
	tx.Commit()
	return
}

// Load all orders with not finished preparation status.
func (base *Repo) LoadPocessing(ctx context.Context) ([]model.Order, error) {

	orders := make([]model.Order, 0)
	query := `
	SELECT  user_id, order_number
		FROM orders 
		WHERE (status = 'NEW' OR status = 'REGISTERED' OR status = 'PROCESSING') AND is_preorder = FALSE
	`
	err := base.master.SelectContext(ctx, &orders, query)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (base *Repo) UpdateStatus(ctx context.Context, order string, status model.Status) (err error) {

	_, err = base.master.ExecContext(ctx, "UPDATE orders SET status = $1 WHERE order_number = $2", status, order)
	if err != nil {
		return fmt.Errorf("can't update orders status, %w", err)
	}

	return
}
