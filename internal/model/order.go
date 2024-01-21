package model

import (
	"time"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

type Order struct {
	UserID     *uuid.UUID
	Onumber    string
	Uploaded   time.Time
	Status     Status
	IsPreOrder bool
	Bonus      *Bonus
}

func NewOrder(userID *uuid.UUID, onumber string, preoreder bool, used *decimal.Decimal, accrual *decimal.Decimal) *Order {

	return &Order{UserID: userID, Onumber: onumber, IsPreOrder: preoreder, Uploaded: time.Now(), Status: Status(0), Bonus: NewBonus(used, accrual)}
}

// Check Luna namber
func (o *Order) IsValid() (isValid bool) {
	err := goluhn.Validate(o.Onumber)
	return err == nil

}
