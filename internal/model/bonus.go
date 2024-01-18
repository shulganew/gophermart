package model

import (
	"github.com/shopspring/decimal"
)

type Bonus struct {
	order   *Order
	used    *decimal.Decimal
	accural *decimal.Decimal
}

func NewBonus(order *Order, used *decimal.Decimal, accural *decimal.Decimal) *Bonus {

	return &Bonus{order: order, used: used, accural: accural}
}
