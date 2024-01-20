package model

import (
	"github.com/shopspring/decimal"
)

type Withdraw struct {
	Onumber   string  `json:"order"`
	Withdrawn float64 `json:"sum"`
}

type Withdrawals struct {
	Onumber   string  `json:"order"`
	Withdrawn float64 `json:"sum"`
	Uploaded  string  `json:"processed_at"`
}

func NewWithdrawals(order string, withdrawn *decimal.Decimal, time string) *Withdrawals {

	w := withdrawn.InexactFloat64()
	return &Withdrawals{Onumber: order, Withdrawn: w, Uploaded: time}
}
