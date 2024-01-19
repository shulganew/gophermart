package model

import (
	"github.com/shopspring/decimal"
)

type Bonus struct {
	Used    *decimal.Decimal
	Accrual *decimal.Decimal
}

func NewBonus(used *decimal.Decimal, accrual *decimal.Decimal) *Bonus {

	return &Bonus{Used: used, Accrual: accrual}
}
