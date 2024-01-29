package model

import (
	"time"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

type Order struct {
	UserID     *uuid.UUID      `db:"user_id"`
	Onumber    string          `db:"onumber"`
	IsPreOrder bool            `db:"is_preorder"`
	Uploaded   time.Time       `db:"uploaded"`
	Status     Status          `db:"status"`
	Withdrawn  decimal.Decimal `db:"withdrawn"`
	Accrual    decimal.Decimal `db:"accrual"`
}

func NewOrder(userID *uuid.UUID, onumber string, preoreder bool, withdrawn decimal.Decimal, accrual decimal.Decimal) *Order {

	return &Order{UserID: userID, Onumber: onumber, IsPreOrder: preoreder, Uploaded: time.Now(), Status: Status(NEW), Withdrawn: withdrawn, Accrual: accrual}
}

// Check Luna namber
func (o *Order) IsValid() (isValid bool) {
	err := goluhn.Validate(o.Onumber)
	return err == nil

}
