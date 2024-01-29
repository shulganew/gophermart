package model

import "github.com/shopspring/decimal"

type UserBalance struct {
	Bonus     float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func NewUserBalance(bonus decimal.Decimal, withdrawn decimal.Decimal) *UserBalance {

	b := bonus.InexactFloat64()
	w := withdrawn.InexactFloat64()
	return &UserBalance{Bonus: b, Withdrawn: w}
}
