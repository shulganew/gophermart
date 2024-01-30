package storage

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

func (base *Repo) GetBonuses(ctx context.Context, userID *uuid.UUID) (accrual decimal.Decimal, err error) {

	err = base.master.GetContext(ctx, &accrual, "SELECT bonuses FROM users where user_id = $1", userID)
	if err != nil {
		return decimal.Zero, err
	}
	return accrual, nil
}

func (base *Repo) SetAccrual(ctx context.Context, order string, accrual decimal.Decimal) (err error) {

	_, err = base.master.ExecContext(ctx, "UPDATE orders SET accrual = $1 WHERE onumber = $2", accrual, order)
	if err != nil {

		return fmt.Errorf("can't update order's accrual during update status %w", err)
	}

	return nil
}

func (base *Repo) AddBonuses(ctx context.Context, userID *uuid.UUID, amount decimal.Decimal) (err error) {

	_, err = base.master.ExecContext(ctx, "UPDATE users SET bonuses = bonuses + $1 WHERE user_id = $2", amount, userID)
	if err != nil {
		return fmt.Errorf("can't update add to user's order accruals to bonuses %w", err)
	}

	return nil
}

// Move user's amount from bonuses to withdrawals.
func (base *Repo) MakeWithdrawn(ctx context.Context, userID *uuid.UUID, amount decimal.Decimal) (err error) {

	tx := base.master.MustBegin()

	_, err = tx.ExecContext(ctx, "UPDATE users SET bonuses = bonuses - $1 WHERE user_id = $2", amount, userID)
	if err != nil {
		return fmt.Errorf("can't make bonuse withdrawn, %w", err)
	}

	//check uses balance after update
	var bonuses decimal.Decimal

	//check uses balance in transaction
	err = tx.GetContext(ctx, &bonuses, "SELECT bonuses FROM users WHERE user_id = $1", userID)
	if err != nil || bonuses.IsNegative() {
		tx.Rollback()
		return fmt.Errorf("error during user's bonuse withdrawn: %w", err)
	}

	//Update user's withdrawals
	_, err = tx.ExecContext(ctx, "UPDATE users SET withdrawals = withdrawals + $1 WHERE user_id = $2", amount, userID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("can't add withdrawns to user, %w", err)
	}

	tx.Commit()

	return
}
