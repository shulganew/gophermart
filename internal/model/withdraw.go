package model

import (
	"github.com/shopspring/decimal"
)

type Withdraw struct {
	OrderNr   string  `json:"order"`
	Withdrawn float64 `json:"sum"`
}

type Withdrawals struct {
	OrderNr   string  `json:"order" db:"order_number"`
	Withdrawn float64 `json:"sum"`
	Uploaded  string  `json:"processed_at"`
}

func NewWithdrawals(order string, withdrawn *decimal.Decimal, time string) *Withdrawals {
	w := withdrawn.InexactFloat64()
	return &Withdrawals{OrderNr: order, Withdrawn: w, Uploaded: time}
}
